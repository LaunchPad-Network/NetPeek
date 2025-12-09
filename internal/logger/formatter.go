package logger

import (
	"bytes"
	"math"
	"strings"
	"sync"

	"github.com/LaunchPad-Network/NetPeek/constant"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type Formatter struct {
	Name string

	preRenderingName                       string
	preGenerateBlankBetweenNameAndPosition string
}

func NewFormatter(name string) *Formatter {
	formatter := &Formatter{
		Name: name,
	}
	formatter.preRender()
	return formatter
}

func (f *Formatter) preRender() {
	name := f.Name
	nameLength := len(name)
	if nameLength > NameMaxLength {
		middleIndex := int(math.Ceil(float64(NameMaxLength) / 2))
		name = name[0:middleIndex] + "..." + name[nameLength+3-middleIndex:nameLength]
	}
	colorRGB := generateNameColorRGB(f.Name)
	f.preRenderingName = colorRGB.Sprint(name)
	f.preGenerateBlankBetweenNameAndPosition = generateBlankBetweenNameAndPosition(len(name))
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	timeStr := entry.Time.Format(constant.TimeFormat)

	logoBackgroundColorRGB, logoFontColorRGB, messageColorRGB := generateLogoBackgroundColorAndLogoFontColorAndMessageColorRGB(entry.Level)

	buf.WriteString(timeStr)
	buf.WriteString("  ")
	buf.WriteString(f.preRenderingName)
	buf.WriteString(f.preGenerateBlankBetweenNameAndPosition)
	buf.WriteString(logoBackgroundColorRGB.Sprintf(" %s", logoFontColorRGB.Sprint(strings.ToUpper(entry.Level.String()[:1]))) + logoBackgroundColorRGB.Sprint(" "))
	buf.WriteString("  ")
	buf.WriteString(messageColorRGB.Sprint(entry.Message))
	if len(entry.Data) != 0 {
		buf.WriteString("  ")
		raw := messageColorRGB.Values()
		fieldsColorRGB := color.RGB(uint8(raw[0]+FiledColorRGBOffset), uint8(raw[1]+FiledColorRGBOffset), uint8(raw[2]+FiledColorRGBOffset))
		buf.WriteString(fieldsColorRGB.Sprint(getFieldsString(entry.Data)))
	}
	buf.WriteString("\n")

	return buf.Bytes(), nil
}
