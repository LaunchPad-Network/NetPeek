package traceroute

import (
	"os/exec"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/logger"

	"github.com/google/shlex"
	"github.com/spf13/viper"
)

var log = logger.New("Traceroute")

func init() {
	tracerouteAutodetect()
}

func tracerouteArgsToString(cmd string, args []string, target []string) string {
	var cmdCombined = append([]string{cmd}, args...)
	cmdCombined = append(cmdCombined, target...)
	return strings.Join(cmdCombined, " ")
}

func tracerouteTryExecute(cmd string, args []string, target []string) ([]byte, error) {
	instance := exec.Command(cmd, append(args, target...)...)
	output, err := instance.CombinedOutput()
	if err == nil {
		return output, nil
	}

	return output, err
}

func tracerouteDetect(cmd string, args []string) bool {
	target := []string{"127.0.0.1"}
	success := false
	if result, err := tracerouteTryExecute(cmd, args, target); err == nil {
		viper.Set("traceroute.binary", cmd)
		viper.Set("traceroute.flags", args)
		success = true
		log.Infof("Traceroute autodetect success: %s\n", tracerouteArgsToString(cmd, args, target))
	} else {
		log.Infof("Traceroute autodetect fail, continuing: %s (%s)\n%s", tracerouteArgsToString(cmd, args, target), err.Error(), result)
	}

	return success
}

func tracerouteAutodetect() {
	if viper.GetString("traceroute.binary") != "" && viper.GetString("traceroute.flags") != "" {
		return
	}

	// Traceroute (custom binary)
	if viper.GetString("traceroute.binary") != "" {
		if tracerouteDetect(viper.GetString("traceroute.binary"), []string{"-q1", "-N32", "-w1"}) {
			return
		}
		if tracerouteDetect(viper.GetString("traceroute.binary"), []string{"-q1", "-w1"}) {
			return
		}
		if tracerouteDetect(viper.GetString("traceroute.binary"), []string{}) {
			return
		}
	}

	// MTR
	if tracerouteDetect("mtr", []string{"-w", "-c1", "-Z1", "-G1", "-b"}) {
		return
	}

	// Traceroute
	if tracerouteDetect("traceroute", []string{"-q1", "-N32", "-w1"}) {
		return
	}
	if tracerouteDetect("traceroute", []string{"-q1", "-w1"}) {
		return
	}
	if tracerouteDetect("traceroute", []string{}) {
		return
	}

	// Unsupported
	viper.Set("traceroute.binary", "")
	viper.Set("traceroute.flags", []string{})
	log.Warn("Traceroute autodetect failed! Traceroute will be disabled")
}

func CallTraceroute(q string) (string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "", ErrEmptyTarget
	}

	args, err := shlex.Split(q)
	if err != nil {
		return "", err
	}

	binary := viper.GetString("traceroute.binary")
	if binary == "" {
		return "", ErrNotSupported
	}

	result, err := tracerouteTryExecute(binary, viper.GetStringSlice("traceroute.flags"), args)
	return string(result), err
}
