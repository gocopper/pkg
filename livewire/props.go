package livewire

import "encoding/json"

func NewProps() *Props {
	return &Props{
		data:  make(map[string]interface{}),
		dirty: make([]string, 0),
	}
}

type Props struct {
	data  map[string]interface{}
	dirty []string
}

func (p *Props) UnmarshalJSON(bytes []byte) error {
	p.dirty = make([]string, 0)
	return json.Unmarshal(bytes, &p.data)
}

func (p *Props) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.data)
}

func (p *Props) Set(k string, v interface{}) {
	p.data[k] = v
	p.dirty = append(p.dirty, k)
}

func (p *Props) GetString(k string) string {
	v, ok := p.data[k]
	if !ok {
		return ""
	}

	vStr, ok := v.(string)
	if !ok {
		return ""
	}

	return vStr
}

func (p *Props) GetInt64(k string) int64 {
	v, ok := p.data[k]
	if !ok {
		return 0
	}

	vInt64, ok := v.(int64)
	if ok {
		return vInt64
	}

	vInt32, ok := v.(int32)
	if ok {
		return int64(vInt32)
	}

	vInt, ok := v.(int)
	if ok {
		return int64(vInt)
	}

	vFloat64, ok := v.(float64)
	if ok {
		return int64(vFloat64)
	}

	vFloat32, ok := v.(float32)
	if ok {
		return int64(vFloat32)
	}

	return 0
}
