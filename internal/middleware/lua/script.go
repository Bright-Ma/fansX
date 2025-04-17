package lua

type Script struct {
	name     string
	function string
}

func (s *Script) Name() string {
	return s.name
}

func (s *Script) Function() string {
	return s.function
}

func NewScript(name string, function string) *Script {
	return &Script{
		name:     name,
		function: function,
	}
}
