package whatsmeow

// import (
// 	"fmt"
// 	"strings"
// 	"time"

// 	waBinary "go.mau.fi/whatsmeow/binary"
// 	"go.mau.fi/whatsmeow/types"
// )

// // BiUtil is a utility for binary attributes.
// // that provides a set of helper functions to get values from binary attributes.
// // It is used to get values from binary attributes.

// type BiUtil struct {
// 	utility *waBinary.Node
// }

// func NewBinaryUtil(node *waBinary.Node) *BiUtil {
// 	return &BiUtil{
// 		utility: node,
// 	}
// }

// func (util *BiUtil) Childrens(tag string) []waBinary.Node {
// 	return util.utility.GetChildrenByTag(tag)
// }

// func (util *BiUtil) String(key string, def string) string {
// 	return findAny(util, key, def)
// }

// // func (util *BiUtil) Int64(key string, def int64) int64 {
// // 	value, ok := util.utility..GetInt64(key, false)
// // 	if !ok {
// // 		return def
// // 	}
// // 	return value
// // }

// // func (util *BiUtil) Int(key string, def int) int {
// // 	value, ok := util.utility.GetInt64(key, false)
// // 	if !ok {
// // 		return def
// // 	}
// // 	return int(value)
// // }

// // func (util *BiUtil) Time(key string, def time.Time) time.Time {
// // 	value, ok := util.utility.GetUnixTime(key, false)
// // 	if !ok {
// // 		return def
// // 	}
// // 	return value
// // }

// // func (util *BiUtil) Bool(key string, def bool) bool {
// // 	value, ok := util.utility.GetBool(key, false)
// // 	if !ok {
// // 		return def
// // 	}
// // 	return value
// // }

// // func (util *BiUtil) Float64(key string) *types.JID {
// // 	value, ok := util.utility.GetJID(key, false)
// // 	if !ok {
// // 		return nil
// // 	}
// // 	return &value
// // }

// /////////////////////////////
// // private methods
// /////////////////////////////

// func keyDotToList(key string) []string {

// 	if strings.Contains(key, ".") {
// 		return strings.Split(key, ".")
// 	}

// 	return []string{key}
// }

// func findAny[T any](util *BiUtil, key string, def T) T {

// 	keys := keyDotToList(key)
// 	var currentData any = def

// 	for _, key := range keys {

// 		/// check if is map
// 		_type := fmt.Sprintf("%T", util.utility.Attrs)

// 		if strings.Contains(_type, "map") {

// 			value, ok := util.utility.Attrs[key]
// 			if ok {
// 				currentData = value
// 			} else {
// 				return def
// 			}

// 		} else {
// 			currentData = def
// 		}
// 	}

// 	value, ok := currentData.(T)
// 	if !ok {
// 		return def
// 	}

// 	return value

// }
