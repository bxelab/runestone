package go_runestone

type Artifact struct {
	Cenotaph  *Cenotaph
	Runestone *Runestone
}

func (a *Artifact) Mint() *RuneId {
	if a.Cenotaph != nil {
		return a.Cenotaph.Mint
	}
	if a.Runestone != nil {
		return a.Runestone.Mint
	}
	return nil
}
