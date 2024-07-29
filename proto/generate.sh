# #!/bin/bash
# cd $(dirname $0)
# set -euo pipefail
# if [[ ! -f "protos.js" ]]; then
# 	echo "Please download the WhatsApp JavaScript modules with protobuf schemas into protos.js first"
# 	exit 1
# fi
# node parse-proto.js
protoc --go_out=. --go_opt=paths=source_relative  */*.proto
# pre-commit run -a


sleep 1