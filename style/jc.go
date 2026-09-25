package style

// Horizontal alignment values (OOXML ST_Jc).
const (
	JcStart          = "start"
	JcCenter         = "center"
	JcEnd            = "end"
	JcBoth           = "both"
	JcMediumKashida  = "mediumKashida"
	JcDistribute     = "distribute"
	JcNumTab         = "numTab"
	JcHighKashida    = "highKashida"
	JcLowKashida     = "lowKashida"
	JcThaiDistribute = "thaiDistribute"
	JcLeft           = "left"
	JcRight          = "right"
	JcJustify        = "both"
)

// Table justification (ST_JcTable).
const (
	JcTableLeft   = "left"
	JcTableCenter = "center"
	JcTableRight  = "right"
	JcTableStart  = "start"
	JcTableEnd    = "end"
)

// Vertical cell alignment (ST_VerticalJc).
const (
	VAlignTop    = "top"
	VAlignCenter = "center"
	VAlignBoth   = "both"
	VAlignBottom = "bottom"
)

// Line spacing rule.
const (
	LineSpacingAuto    = "auto"
	LineSpacingExact   = "exact"
	LineSpacingAtLeast = "atLeast"
)

// Number formats (ST_NumberFormat).
const (
	NumberDecimal      = "decimal"
	NumberUpperRoman   = "upperRoman"
	NumberLowerRoman   = "lowerRoman"
	NumberUpperLetter  = "upperLetter"
	NumberLowerLetter  = "lowerLetter"
	NumberBullet       = "bullet"
	NumberOrdinal      = "ordinal"
	NumberCardinalText = "cardinalText"
	NumberOrdinalText  = "ordinalText"
	NumberHex          = "hex"
	NumberChicago      = "chicago"
	NumberNone         = "none"
	NumberDecimalZero  = "decimalZero"

	NumberIdeographDigital            = "ideographDigital"
	NumberJapaneseCounting            = "japaneseCounting"
	NumberAiueo                       = "aiueo"
	NumberIroha                       = "iroha"
	NumberDecimalFullWidth            = "decimalFullWidth"
	NumberDecimalHalfWidth            = "decimalHalfWidth"
	NumberJapaneseLegal               = "japaneseLegal"
	NumberJapaneseDigitalTenThousand  = "japaneseDigitalTenThousand"
	NumberDecimalEnclosedCircle       = "decimalEnclosedCircle"
	NumberDecimalFullWidth2           = "decimalFullWidth2"
	NumberAiueoFullWidth              = "aiueoFullWidth"
	NumberIrohaFullWidth              = "irohaFullWidth"
	NumberGanada                      = "ganada"
	NumberChosung                     = "chosung"
	NumberDecimalEnclosedFullStop     = "decimalEnclosedFullstop"
	NumberDecimalEnclosedParen        = "decimalEnclosedParen"
	NumberDecimalEnclosedCircleChinese = "decimalEnclosedCircleChinese"
	NumberIdeographEnclosedCircle     = "ideographEnclosedCircle"
	NumberIdeographTraditional        = "ideographTraditional"
	NumberIdeographZodiac             = "ideographZodiac"
	NumberIdeographZodiacTraditional  = "ideographZodiacTraditional"
	NumberTaiwaneseCounting           = "taiwaneseCounting"
	NumberIdeographLegalTraditional   = "ideographLegalTraditional"
	NumberTaiwaneseCountingThousand   = "taiwaneseCountingThousand"
	NumberTaiwaneseDigital            = "taiwaneseDigital"
	NumberChineseCounting             = "chineseCounting"
	NumberChineseLegalSimplified      = "chineseLegalSimplified"
	NumberChineseCountingThousand     = "chineseCountingThousand"
	NumberKoreanDigital               = "koreanDigital"
	NumberKoreanCounting              = "koreanCounting"
	NumberKoreanLegal                 = "koreanLegal"
	NumberKoreanDigital2              = "koreanDigital2"
	NumberVietnameseCounting          = "vietnameseCounting"
	NumberRussianLower                = "russianLower"
	NumberRussianUpper                = "russianUpper"
	NumberNumberInDash                = "numberInDash"
	NumberHebrew1                     = "hebrew1"
	NumberHebrew2                     = "hebrew2"
	NumberArabicAlpha                 = "arabicAlpha"
	NumberArabicAbjad                 = "arabicAbjad"
	NumberHindiVowels                 = "hindiVowels"
	NumberHindiConsonants             = "hindiConsonants"
	NumberHindiNumbers                = "hindiNumbers"
	NumberHindiCounting               = "hindiCounting"
	NumberThaiLetters                 = "thaiLetters"
	NumberThaiNumbers                 = "thaiNumbers"
	NumberThaiCounting                = "thaiCounting"
)

// Border styles (PHPWord SimpleType\Border).
const (
	BorderSingle              = "single"
	BorderDashDotStroked      = "dashDotStroked"
	BorderDashed              = "dashed"
	BorderDashSmallGap        = "dashSmallGap"
	BorderDotDash             = "dotDash"
	BorderDotDotDash          = "dotDotDash"
	BorderDotted              = "dotted"
	BorderDouble              = "double"
	BorderDoubleWave          = "doubleWave"
	BorderInset               = "inset"
	BorderNil                 = "nil"
	BorderNone                = "none"
	BorderOutset              = "outset"
	BorderThick               = "thick"
	BorderThickThinLargeGap   = "thickThinLargeGap"
	BorderThickThinMediumGap  = "thickThinMediumGap"
	BorderThickThinSmallGap   = "thickThinSmallGap"
	BorderThinThickLargeGap   = "thinThickLargeGap"
	BorderThinThickMediumGap  = "thinThickMediumGap"
	BorderThinThickSmallGap   = "thinThickSmallGap"
	BorderThinThickThinLarge  = "thinThickThinLargeGap"
	BorderThinThickThinMedium = "thinThickThinMediumGap"
	BorderThinThickThinSmall  = "thinThickThinSmallGap"
	BorderThreeDEmboss        = "threeDEmboss"
	BorderThreeDEngrave       = "threeDEngrave"
	BorderTriple              = "triple"
	BorderWave                = "wave"
)

// Document protection modes (PHPWord SimpleType\DocProtect).
const (
	DocProtectNone           = "none"
	DocProtectReadOnly       = "readOnly"
	DocProtectComments       = "comments"
	DocProtectTrackedChanges = "trackedChanges"
	DocProtectForms          = "forms"
)

// Table width units (PHPWord SimpleType\TblWidth).
const (
	TblWidthNil     = "nil"
	TblWidthAuto    = "auto"
	TblWidthPercent = "pct"
	TblWidthTwip    = "dxa"
)

// Text alignment relative to line (PHPWord SimpleType\TextAlignment).
const (
	TextAlignTop      = "top"
	TextAlignCenter   = "center"
	TextAlignBaseline = "baseline"
	TextAlignBottom   = "bottom"
	TextAlignAuto     = "auto"
)

// Zoom presets (PHPWord SimpleType\Zoom).
const (
	ZoomNone     = "none"
	ZoomFullPage = "fullPage"
	ZoomBestFit  = "bestFit"
	ZoomTextFit  = "textFit"
)
