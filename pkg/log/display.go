package log

import "github.com/mgutz/ansi"

const banner = `
  __ _  ___   ___ 
 / _  |/ _ \ / __|
| (_| | (_) | (__ 
 \__, |\___/ \___|
 |___/ 

`

func DisplayGoc() {
	stdout.Write([]byte(ansi.Color(banner, "cyan+b")))
}
