package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Ld(dark, light lipgloss.Color) lipgloss.Color {
	if lipgloss.HasDarkBackground() {
		return dark
	}
	return light
}

var (
	White           = lipgloss.Color("#f6eeee")
	TorchRed        = lipgloss.Color("#ff2d25")
	DeepRed         = lipgloss.Color("#a81c17")
	AlizarinCrimson = lipgloss.Color("#e9362f")
	Thunderbird     = lipgloss.Color("#cf2b25")
	Carnation       = lipgloss.Color("#ff746d")
	Mandy           = lipgloss.Color("#e86a63")
	ChestnutRose    = lipgloss.Color("#d66a64")
	BurntSienna     = lipgloss.Color("#f0746d")
	OldRose         = lipgloss.Color("#c58d8b")
	Sandstone       = lipgloss.Color("#a18f8d")
	Chicago         = lipgloss.Color("#746968")
	SuccessGreen    = Ld(lipgloss.Color("#69d98a"), lipgloss.Color("#00875f"))
)

var (
	outerBorderColor = White
	redColor         = TorchRed
	sepColor         = White
	textColor        = White
	mutedColor       = Sandstone
)

var (
	// Structural borders remain white on dark terminals. Active controls use
	// InnerBorderStyle so the selected item is unmistakably Torch Red.
	BorderStyle      = lipgloss.NewStyle().Foreground(outerBorderColor)
	InnerBorderStyle = lipgloss.NewStyle().Foreground(redColor)

	BadgeBorderStyle    = lipgloss.NewStyle().Foreground(outerBorderColor)
	BadgeStyle          = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeaderTextStyle     = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeroCaptionStyle    = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeaderMetaStyle     = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeaderURLStyle      = lipgloss.NewStyle().Foreground(TorchRed)
	HeaderURLHoverStyle = lipgloss.NewStyle().Underline(true).Foreground(White)

	TitleStyle            = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	TitleIconStyle        = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	DescStyle             = lipgloss.NewStyle().Foreground(mutedColor)
	PrimaryTextStyle      = lipgloss.NewStyle().Foreground(textColor)
	SecondaryTextStyle    = lipgloss.NewStyle().Foreground(mutedColor)
	AccentTextStyle       = lipgloss.NewStyle().Foreground(redColor)
	RowLabelStyle         = lipgloss.NewStyle().Foreground(textColor)
	RowArrowStyle         = lipgloss.NewStyle().Foreground(redColor)
	RowValueStyle         = lipgloss.NewStyle().Foreground(mutedColor)
	FocusedRowLabelStyle  = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	FocusedRowArrowStyle  = lipgloss.NewStyle().Foreground(textColor)
	FocusedRowValueStyle  = lipgloss.NewStyle().Foreground(textColor)
	HoverRowLabelStyle    = lipgloss.NewStyle().Foreground(redColor)
	HoverRowArrowStyle    = lipgloss.NewStyle().Foreground(textColor)
	HoverRowValueStyle    = lipgloss.NewStyle().Foreground(textColor)
	SummaryLabelStyle     = lipgloss.NewStyle().Foreground(textColor)
	SummaryArrowStyle     = lipgloss.NewStyle().Foreground(redColor)
	SummaryValueStyle     = lipgloss.NewStyle().Foreground(textColor)
	SummaryMetaStyle      = lipgloss.NewStyle().Foreground(mutedColor)
	SectionStyle          = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	SectionIconStyle      = lipgloss.NewStyle().Foreground(textColor)
	KeyStyle              = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	RowIconStyle          = lipgloss.NewStyle().Foreground(redColor)
	IconStyle             = lipgloss.NewStyle().Foreground(textColor)
	LabelStyle            = lipgloss.NewStyle().Foreground(textColor)
	MutedStyle            = lipgloss.NewStyle().Foreground(mutedColor)
	PathStyle             = lipgloss.NewStyle().Foreground(DeepRed)
	ModeStyle             = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	SelectedModeStyle     = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	MnemonicStyle         = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(redColor)
	SelectedMnemonicStyle = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(textColor)
	PromptStyle           = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	WarnStyle             = lipgloss.NewStyle().Foreground(BurntSienna)
	ErrorStyle            = lipgloss.NewStyle().Foreground(TorchRed)
	SuccessStyle          = lipgloss.NewStyle().Foreground(textColor)
	CompleteStyle         = lipgloss.NewStyle().Bold(true).Foreground(SuccessGreen)
	InfoStyle             = lipgloss.NewStyle().Foreground(BurntSienna)
	ArtistStyle           = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	AlbumStyle            = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	ActiveRowStyle        = lipgloss.NewStyle().Foreground(redColor)
	SepStyle              = lipgloss.NewStyle().Foreground(sepColor)
)

// HeroWordmarkRed is the solid face of the ASCII block wordmark: every full
// block glyph renders in the primary Last.fm red.
var HeroWordmarkRed = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff2c25"))

// HeroWordmarkShadow is the dimensional/outline backing of the ASCII block
// wordmark: every box-drawing or shadow glyph behind the lettering renders in
// the dim hint tone (Sandstone) so the red face stays dominant.
var HeroWordmarkShadow = lipgloss.NewStyle().Foreground(Sandstone)

// HeroWordmarkLine paints one ASCII wordmark row with deterministic
// glyph-based two-tone coloring: full block cells turn Last.fm red, outline
// cells turn the dim hint tone, and spaces stay uncolored. Styling is per
// rune, never per row, and the emitted ANSI has zero display width, so styled
// lines keep the exact geometry of the plain artwork.
func HeroWordmarkLine(line string) string {
	var out strings.Builder
	for _, r := range line {
		switch r {
		case '█':
			out.WriteString(HeroWordmarkRed.Render(string(r)))
		case '╔', '╗', '╚', '╝', '═', '║':
			out.WriteString(HeroWordmarkShadow.Render(string(r)))
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
