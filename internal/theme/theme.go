package theme

import "github.com/charmbracelet/lipgloss"

func Ld(dark, light lipgloss.Color) lipgloss.Color {
	if lipgloss.HasDarkBackground() {
		return dark
	}
	return light
}

var (
	White           = lipgloss.Color("#ffffff")
	TorchRed        = lipgloss.Color("#f8211c")
	DeepRed         = lipgloss.Color("#991411")
	AlizarinCrimson = lipgloss.Color("#e11e19")
	Thunderbird     = lipgloss.Color("#cc1b17")
	Carnation       = lipgloss.Color("#fa645d")
	Mandy           = lipgloss.Color("#e25951")
	ChestnutRose    = lipgloss.Color("#d0524a")
	BurntSienna     = lipgloss.Color("#f06059")
	OldRose         = lipgloss.Color("#be7776")
	Sandstone       = lipgloss.Color("#736765")
	Chicago         = lipgloss.Color("#565554")
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

	BadgeBorderStyle = lipgloss.NewStyle().Foreground(outerBorderColor)
	BadgeStyle       = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeaderTextStyle  = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	HeaderMetaStyle  = lipgloss.NewStyle().Bold(true).Foreground(DeepRed)

	TitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	TitleIconStyle   = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	DescStyle        = lipgloss.NewStyle().Foreground(mutedColor)
	SectionStyle     = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	SectionIconStyle = lipgloss.NewStyle().Foreground(textColor)
	KeyStyle         = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	RowIconStyle     = lipgloss.NewStyle().Foreground(redColor)
	IconStyle        = lipgloss.NewStyle().Foreground(textColor)
	LabelStyle       = lipgloss.NewStyle().Foreground(textColor)
	MutedStyle       = lipgloss.NewStyle().Foreground(mutedColor)
	PathStyle        = lipgloss.NewStyle().Foreground(DeepRed)
	ModeStyle        = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	PromptStyle      = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	WarnStyle        = lipgloss.NewStyle().Foreground(BurntSienna)
	ErrorStyle       = lipgloss.NewStyle().Foreground(TorchRed)
	SuccessStyle     = lipgloss.NewStyle().Foreground(textColor)
	InfoStyle        = lipgloss.NewStyle().Foreground(BurntSienna)
	ArtistStyle      = lipgloss.NewStyle().Bold(true).Foreground(redColor)
	AlbumStyle       = lipgloss.NewStyle().Bold(true).Foreground(textColor)
	ActiveRowStyle   = lipgloss.NewStyle().Foreground(redColor)
	SepStyle         = lipgloss.NewStyle().Foreground(sepColor)
)
