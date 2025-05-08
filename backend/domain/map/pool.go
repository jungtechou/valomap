package domain

type MapPool struct {
	Maps []Map
}

func (p *MapPool) AvailableMaps(banList []string) []Map {
	banned := make(map[string]bool)
	for _, id := range banList {
		banned[id] = true
	}

	var available []Map
	for _, m := range p.Maps {
		if !banned[m.UUID] {
			available = append(available, m)
		}
	}
	return available
}
