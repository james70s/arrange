package config

var (
	C *Config
)

type Config struct {
	Color *color
}

type color struct {
	Black            int `toml:"black"`
	White            int `toml:"white"`
	Red              int `toml:"red"`
	Purple           int `toml:"purple"`
	Logo             int `toml:"logo"`
	Yellow           int `toml:"yellow"`
	Green            int `toml:"green"`
	Menu             int `toml:"menu"`
	MyNick           int `toml:"my_nick"`
	OtherNickDefault int `toml:"other_nick_default"`
	Timestamp        int `toml:"timestamp"`
	MyText           int `toml:"my_text"`
	Header           int `toml:"header"`
	QueryHeader      int `toml:"query_header"`
	CurrentInputView int `toml:"current_input_view"`
	Notice           int `toml:"notice"`
	Action           int `toml:"action"`
}

func Default() *Config {
	return &Config{
		Color: &color{
			Notice:           219,
			Action:           118,
			Black:            0,
			White:            15,
			Red:              160,
			Purple:           92,
			Logo:             75,
			Yellow:           11,
			Green:            119,
			Menu:             209,
			MyNick:           119,
			OtherNickDefault: 14,
			Timestamp:        247,
			MyText:           129,
			Header:           57,
			QueryHeader:      11,
			CurrentInputView: 215,
		},
		// Time: &time{
		// 	MessageFormat: "15:04",
		// 	NoticeFormat:  "02 Jan 06 15:04 MST",
		// 	MenuFormat:    "03:04:05 PM",
		// },
	}
}

func init() {
	C = Default()
}
