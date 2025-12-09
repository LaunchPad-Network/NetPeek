package logger

import "github.com/gookit/color"

const (
	NameMaxLength     = 22
	PositionMaxLength = 20
)

const FiledColorRGBOffset = -25

var (
	MessageFontColorDebug = color.RGB(41, 153, 153)
	MessageFontColorInfo  = color.RGB(171, 192, 35)
	MessageFontColorWarn  = color.RGB(187, 181, 41)
	MessageFontColorError = color.RGB(255, 107, 104)
	MessageFontColorFatal = color.RGB(237, 116, 109)

	LogoBackgroundColorDebug = color.RGB(48, 93, 120, true)
	LogoBackgroundColorInfo  = color.RGB(106, 135, 89, true)
	LogoBackgroundColorWarn  = color.RGB(187, 181, 41, true)
	LogoBackgroundColorError = color.RGB(207, 91, 86, true)
	LogoBackgroundColorFatal = color.RGB(139, 60, 60, true)

	LogoFontColorDebug = color.RGB(187, 187, 187)
	LogoFontColorInfo  = color.RGB(233, 245, 230)
	LogoFontColorWarn  = color.RGB(0, 0, 0)
	LogoFontColorError = color.RGB(0, 0, 0)
	LogoFontColorFatal = color.RGB(0, 0, 0)
)

var (
	ColorCharABYZ = color.RGB(164, 173, 115)
	ColorCharCD12 = color.RGB(154, 151, 58)
	ColorCharEF34 = color.RGB(168, 139, 187)
	ColorCharGH56 = color.RGB(120, 145, 197)
	ColorCharIJ78 = color.RGB(78, 171, 241)
	ColorCharKL90 = color.RGB(221, 175, 209)
	ColorCharMN   = color.RGB(183, 150, 108)
	ColorCharOP   = color.RGB(103, 214, 193)
	ColorCharQR   = color.RGB(189, 174, 202)
	ColorCharST   = color.RGB(205, 150, 100)
	ColorCharUV   = color.RGB(198, 125, 130)
	ColorCharWX   = color.RGB(216, 203, 224)
)
