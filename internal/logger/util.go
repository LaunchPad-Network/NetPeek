package logger

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/misc/crypto/md5"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

func generateNameColorRGB(name string) color.RGBColor {
	md5 := md5.StringHex(name)
	md5PrefixUpper := strings.ToUpper(md5[:1])

	if strings.ContainsAny(md5PrefixUpper, "ABYZ") {
		return ColorCharABYZ
	} else if strings.ContainsAny(md5PrefixUpper, "CD12") {
		return ColorCharCD12
	} else if strings.ContainsAny(md5PrefixUpper, "EF34") {
		return ColorCharEF34
	} else if strings.ContainsAny(md5PrefixUpper, "GH56") {
		return ColorCharGH56
	} else if strings.ContainsAny(md5PrefixUpper, "IJ78") {
		return ColorCharIJ78
	} else if strings.ContainsAny(md5PrefixUpper, "KL90") {
		return ColorCharKL90
	} else if strings.ContainsAny(md5PrefixUpper, "MN") {
		return ColorCharMN
	} else if strings.ContainsAny(md5PrefixUpper, "OP") {
		return ColorCharOP
	} else if strings.ContainsAny(md5PrefixUpper, "QR") {
		return ColorCharQR
	} else if strings.ContainsAny(md5PrefixUpper, "ST") {
		return ColorCharST
	} else if strings.ContainsAny(md5PrefixUpper, "UV") {
		return ColorCharUV
	} else if strings.ContainsAny(md5PrefixUpper, "WX") {
		return ColorCharWX
	} else {
		return ColorCharABYZ
	}
}

func generateLogoBackgroundColorAndLogoFontColorAndMessageColorRGB(level logrus.Level) (color.RGBColor, color.RGBColor, color.RGBColor) {
	switch level {
	case logrus.DebugLevel:
		return LogoBackgroundColorDebug, LogoFontColorDebug, MessageFontColorDebug
	case logrus.InfoLevel:
		return LogoBackgroundColorInfo, LogoFontColorInfo, MessageFontColorInfo
	case logrus.WarnLevel:
		return LogoBackgroundColorWarn, LogoFontColorWarn, MessageFontColorWarn
	case logrus.ErrorLevel:
		return LogoBackgroundColorError, LogoFontColorError, MessageFontColorError
	case logrus.FatalLevel:
		return LogoBackgroundColorFatal, LogoFontColorFatal, MessageFontColorFatal
	default:
		return LogoBackgroundColorDebug, LogoFontColorDebug, MessageFontColorDebug
	}
}

func generateBlankBetweenNameAndPosition(nameLength int) string {
	return strings.Repeat(" ", NameMaxLength-nameLength+2)
}

func getFieldsString(f logrus.Fields) string {
	length := len(f)
	if length == 0 {
		return "{\t}"
	}

	keys := make([]string, 0, length)
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb := strings.Builder{}
	sb.WriteString("{ ")
	for i, k := range keys {
		sb.WriteString(fmt.Sprintf("%s:\t%v", k, f[k]))
		if i < length-1 {
			sb.WriteString(",\t")
		}
	}
	sb.WriteString(" }")
	return sb.String()
}
