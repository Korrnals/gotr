package pdf

import _ "embed"

//go:embed fonts/NotoSans-Regular.ttf
var notoSansRegular []byte

//go:embed fonts/NotoSans-Bold.ttf
var notoSansBold []byte

// FontFamily is the family name registered with fpdf.
const FontFamily = "NotoSans"
