// Copyright (c) 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	mathRand "math/rand/v2"

	"github.com/google/uuid"
	"go.mau.fi/util/dbutil"
	"go.mau.fi/util/random"

	"go.mau.fi/whatsmeow/proto/waAdv"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore/upgrades"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/util/keys"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mau.fi/whatsmeow/util/sqlutil"
)

// Container is a wrapper for a SQL database that can contain multiple whatsmeow sessions.
type Container struct {
	db     *dbutil.Database
	log    waLog.Logger
	LIDMap *CachedLIDMap

	// default namespace for devices in this container.
	Namespace string
}

var _ store.DeviceContainer = (*Container)(nil)

// New connects to the given SQL database and wraps it in a Container.
//
// Only SQLite and Postgres are currently fully supported.
//
// The logger can be nil and will default to a no-op logger.
//
// When using SQLite, it's strongly recommended to enable foreign keys by adding `?_foreign_keys=true`:
//
//	container, err := sqlstore.New(context.Background(), "sqlite3", "file:yoursqlitefile.db?_foreign_keys=on", nil)
func New(ctx context.Context, dialect, address string, log waLog.Logger) (*Container, error) {
	db, err := sql.Open(dialect, address)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	container := NewWithDB(db, dialect, log)
	err = container.Upgrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade database: %w", err)
	}
	return container, nil
}

// NewWithDB wraps an existing SQL connection in a Container.
//
// Only SQLite and Postgres are currently fully supported.
//
// The logger can be nil and will default to a no-op logger.
//
// When using SQLite, it's strongly recommended to enable foreign keys by adding `?_foreign_keys=true`:
//
//	db, err := sql.Open("sqlite3", "file:yoursqlitefile.db?_foreign_keys=on")
//	if err != nil {
//	    panic(err)
//	}
//	container := sqlstore.NewWithDB(db, "sqlite3", nil)
//
// This method does not call Upgrade automatically like New does, so you must call it yourself:
//
//	container := sqlstore.NewWithDB(...)
//	err := container.Upgrade()
func NewWithDB(db *sql.DB, dialect string, log waLog.Logger) *Container {
	wrapped, err := dbutil.NewWithDB(db, dialect)
	if err != nil {
		// This will only panic if the dialect is invalid
		panic(err)
	}
	wrapped.UpgradeTable = upgrades.Table
	wrapped.VersionTable = "whatsmeow_version"
	return NewWithWrappedDB(wrapped, log)
}

// NewWithDBAndNamespace wraps an existing SQL connection in a Container with a specific namespace.
func NewWithDBAndNamespace(db *sql.DB, dialect string, log waLog.Logger, namespace string) *Container {
	container := NewWithDB(db, dialect, log)
	container.Namespace = namespace
	return container
}

func NewWithWrappedDB(wrapped *dbutil.Database, log waLog.Logger) *Container {
	if log == nil {
		log = waLog.Noop
	}
	return &Container{
		db:        wrapped,
		log:       log,
		LIDMap:    NewCachedLIDMap(wrapped),
		Namespace: "",
	}
}

// Upgrade upgrades the database from the current to the latest version available.
func (c *Container) Upgrade(ctx context.Context) error {
	if c.db.Dialect == dbutil.SQLite {
		var foreignKeysEnabled bool
		err := c.db.QueryRow(ctx, "PRAGMA foreign_keys").Scan(&foreignKeysEnabled)
		if err != nil {
			return fmt.Errorf("failed to check if foreign keys are enabled: %w", err)
		} else if !foreignKeysEnabled {
			return fmt.Errorf("foreign keys are not enabled")
		}
	}

	return c.db.Upgrade(ctx)
}

const getAllDevicesQuery = `
SELECT jid, lid, registration_id, noise_key, identity_key,
       signed_pre_key, signed_pre_key_id, signed_pre_key_sig,
       adv_key, adv_details, adv_account_sig, adv_account_sig_key, adv_device_sig,
       platform, business_name, push_name, facebook_uuid, lid_migration_ts, external_id, namespace
FROM whatsmeow_device
`

const getDeviceQuery = getAllDevicesQuery + " WHERE jid=$1"
const getDeviceByExternalID = getAllDevicesQuery + " WHERE external_id=$1"

func (c *Container) scanDevice(row dbutil.Scannable) (*store.Device, error) {
	var device store.Device
	device.Log = c.log
	device.SignedPreKey = &keys.PreKey{}
	var noisePriv, identityPriv, preKeyPriv, preKeySig []byte
	var account waAdv.ADVSignedDeviceIdentity
	var fbUUID uuid.NullUUID

	err := row.Scan(
		&device.ID, &device.LID, &device.RegistrationID, &noisePriv, &identityPriv,
		&preKeyPriv, &device.SignedPreKey.KeyID, &preKeySig,
		&device.AdvSecretKey, &account.Details, &account.AccountSignature, &account.AccountSignatureKey, &account.DeviceSignature,
		&device.Platform, &device.BusinessName, &device.PushName, &fbUUID, &device.LIDMigrationTimestamp,
		&device.ExternalID, &device.Namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan session: %w", err)
	} else if len(noisePriv) != 32 || len(identityPriv) != 32 || len(preKeyPriv) != 32 || len(preKeySig) != 64 {
		return nil, ErrInvalidLength
	}

	device.NoiseKey = keys.NewKeyPairFromPrivateKey(*(*[32]byte)(noisePriv))
	device.IdentityKey = keys.NewKeyPairFromPrivateKey(*(*[32]byte)(identityPriv))
	device.SignedPreKey.KeyPair = *keys.NewKeyPairFromPrivateKey(*(*[32]byte)(preKeyPriv))
	device.SignedPreKey.Signature = (*[64]byte)(preKeySig)
	device.Account = &account
	device.FacebookUUID = fbUUID.UUID

	c.initializeDevice(&device)

	return &device, nil
}

// GetAllDevices finds all the devices in the database.
func (c *Container) GetAllDevices(ctx context.Context) ([]*store.Device, error) {
	res, err := c.db.Query(ctx, getAllDevicesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	sessions := make([]*store.Device, 0)
	for res.Next() {
		sess, scanErr := c.scanDevice(res)
		if scanErr != nil {
			return sessions, scanErr
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// GetAllDevicesByNamespace finds all the devices in the database that belong to the specified namespace.
func (c *Container) GetAllDevicesByNamespace(ctx context.Context, namespace string) ([]*store.Device, error) {
	query := getAllDevicesQuery + " WHERE namespace=$1"
	res, err := c.db.Query(ctx, query, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	sessions := make([]*store.Device, 0)
	for res.Next() {
		sess, scanErr := c.scanDevice(res)
		if scanErr != nil {
			return sessions, scanErr
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// GetFirstDevice is a convenience method for getting the first device in the store. If there are
// no devices, then a new device will be created. You should only use this if you don't want to
// have multiple sessions simultaneously.
func (c *Container) GetFirstDevice(ctx context.Context) (*store.Device, error) {
	devices, err := c.GetAllDevices(ctx)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return c.NewDevice(), nil
	} else {
		return devices[0], nil
	}
}

// GetDevice finds the device with the specified JID in the database.
//
// If the device is not found, nil is returned instead.
//
// Note that the parameter usually must be an AD-JID.
func (c *Container) GetDevice(ctx context.Context, jid types.JID) (*store.Device, error) {
	sess, err := c.scanDevice(c.db.QueryRow(ctx, getDeviceQuery, jid))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return sess, err
}

// GetDeviceByExternalID finds the device with the specified external ID in the database.
func (c *Container) GetDeviceByExternalID(ctx context.Context, externalID string) (*store.Device, error) {
	sess, err := c.scanDevice(c.db.QueryRow(ctx, getDeviceByExternalID, externalID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return sess, err
}

const (
	insertDeviceQuery = `
		INSERT INTO whatsmeow_device (jid, lid, registration_id, noise_key, identity_key,
									  signed_pre_key, signed_pre_key_id, signed_pre_key_sig,
									  adv_key, adv_details, adv_account_sig, adv_account_sig_key, adv_device_sig,
									  platform, business_name, push_name, facebook_uuid, lid_migration_ts, external_id, namespace)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (jid) DO UPDATE
			SET lid=excluded.lid,
				platform=excluded.platform,
				business_name=excluded.business_name,
				push_name=excluded.push_name,
				lid_migration_ts=excluded.lid_migration_ts
	`

	updateDeviceNameSpaces = `UPDATE whatsmeow_device SET namespace=$1 WHERE namespace=$2`
	deleteDeviceQuery      = `DELETE FROM whatsmeow_device WHERE jid=$1`
)

// NewDevice creates a new device in this database.
//
// No data is actually stored before Save is called. However, the pairing process will automatically
// call Save after a successful pairing, so you most likely don't need to call it yourself.
func (c *Container) NewDevice() *store.Device {
	device := &store.Device{
		Log:       c.log,
		Container: c,

		NoiseKey:       keys.NewKeyPair(),
		IdentityKey:    keys.NewKeyPair(),
		RegistrationID: mathRand.Uint32(),
		AdvSecretKey:   random.Bytes(32),
		Namespace:      c.Namespace,
	}
	device.SignedPreKey = device.IdentityKey.CreateSignedPreKey(1)
	return device
}

// NewDeviceWithExternalID creates a new device in this databasw with the given ExternalID.
//
// No data is actually stored before Save is called. However, the pairing process will automatically
// call Save after a successful pairing, so you most likely don't need to call it yourself.
func (c *Container) NewDeviceWithExternalID(externalID string) *store.Device {
	device := c.NewDevice()
	device.ExternalID = externalID
	return device
}

// NewDeviceWithNamespace creates a new device in this database with the given ExternalID and Namespace.
//
// No data is actually stored before Save is called. However, the pairing process will automatically
// call Save after a successful pairing, so you most likely don't need to call it yourself.
func (c *Container) NewDeviceWithNamespace(namespace string) *store.Device {
	device := c.NewDevice()
	device.Namespace = namespace
	return device
}

// NewDeviceWith creates a new device in this database with the given ExternalID and Namespace.
//
// No data is actually stored before Save is called. However, the pairing process will automatically
// call Save after a successful pairing, so you most likely don't need to call it yourself.
func (c *Container) NewDeviceWith(externalID string, namespace string) *store.Device {
	device := c.NewDeviceWithExternalID(externalID)
	device.Namespace = namespace
	return device
}

// ChangeDevicesNamespace changes the namespace of all devices in the database to the specified namespace.
// This is useful when you want to migrate from one static shard (custer-1) to another shard (cluster-2).
func (c *Container) ChangeDevicesNamespace(ctx context.Context, from string, to string, options *sqlutil.SqlLimitOptions) error {

	query := updateDeviceNameSpaces

	if options != nil {

		sqlFragment := options.ToSQL(c.db.Dialect)

		query += fmt.Sprintf(` AND jid IN (
				SELECT jid FROM whatsmeow_device 
				WHERE namespace = %s
				%s
			)`, from, sqlFragment)

	}

	_, err := c.db.Exec(ctx, query, from, to)
	if err != nil {
		return fmt.Errorf("failed to change devices namespace: %w", err)
	}
	return nil
}

// ErrDeviceIDMustBeSet is the error returned by PutDevice if you try to save a device before knowing its JID.
var ErrDeviceIDMustBeSet = errors.New("device JID must be known before accessing database")

// Close will close the container's database
func (c *Container) Close() error {
	if c != nil && c.db != nil {
		return c.db.Close()
	}
	return nil
}

// PutDevice stores the given device in this database. This should be called through Device.Save()
// (which usually doesn't need to be called manually, as the library does that automatically when relevant).
func (c *Container) PutDevice(ctx context.Context, device *store.Device) error {
	if device.ID == nil {
		return ErrDeviceIDMustBeSet
	}
	_, err := c.db.Exec(ctx, insertDeviceQuery,
		device.ID, device.LID, device.RegistrationID, device.NoiseKey.Priv[:], device.IdentityKey.Priv[:],
		device.SignedPreKey.Priv[:], device.SignedPreKey.KeyID, device.SignedPreKey.Signature[:],
		device.AdvSecretKey, device.Account.Details, device.Account.AccountSignature, device.Account.AccountSignatureKey, device.Account.DeviceSignature,
		device.Platform, device.BusinessName, device.PushName, uuid.NullUUID{UUID: device.FacebookUUID, Valid: device.FacebookUUID != uuid.Nil},
		device.LIDMigrationTimestamp,
		device.ExternalID,
		device.Namespace,
	)

	if !device.Initialized {
		c.initializeDevice(device)
	}
	return err
}

func (c *Container) initializeDevice(device *store.Device) {
	innerStore := NewSQLStore(c, *device.ID)
	device.Identities = innerStore
	device.Sessions = innerStore
	device.PreKeys = innerStore
	device.SenderKeys = innerStore
	device.AppStateKeys = innerStore
	device.AppState = innerStore
	device.Contacts = innerStore
	device.ChatSettings = innerStore
	device.MsgSecrets = innerStore
	device.PrivacyTokens = innerStore
	device.EventBuffer = innerStore
	device.LIDs = c.LIDMap
	device.Container = c
	device.Initialized = true
}

// DeleteDevice deletes the given device from this database. This should be called through Device.Delete()
func (c *Container) DeleteDevice(ctx context.Context, store *store.Device) error {
	if store.ID == nil {
		return ErrDeviceIDMustBeSet
	}
	_, err := c.db.Exec(ctx, deleteDeviceQuery, store.ID)
	return err
}
