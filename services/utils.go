package services

import "os"

type Utils struct{}

func NewUtils() *Utils {
	return &Utils{}
}

func (utils *Utils) ExtractPortFromCmdline(pid string) string {
	cmdline, err := os.ReadFile("/proc/" + pid + "/cmdline")
	if err != nil {
		return "N/A"
	}

	args := string(cmdline)
	for i := 0; i < len(args)-1; i++ {
		if args[i] == '-' && args[i+1] == 'p' && i+2 < len(args) {
			i += 3
			port := ""
			for i < len(args) && args[i] != 0 {
				port += string(args[i])
				i++
			}

			if port != "" {
				return port
			}
		}
	}

	return "N/A"
}
