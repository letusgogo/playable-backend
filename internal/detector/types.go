package detector

import "time"

type Area struct {
	Clue   string  `mapstructure:"clue"`
	X      float64 `mapstructure:"x"`
	Y      float64 `mapstructure:"y"`
	Width  float64 `mapstructure:"width"`
	Height float64 `mapstructure:"height"`
}

type Reco struct {
	Method string   `mapstructure:"method"`
	Matchs []string `mapstructure:"matchs"`
}

type Stage struct {
	Number   int           `mapstructure:"number"`
	Interval time.Duration `mapstructure:"interval"`
	Area     Area          `mapstructure:"area"`
	Reco     Reco          `mapstructure:"reco"`
}
