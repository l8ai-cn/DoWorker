package skill

import "context"

func (p *fakePackager) DeletePackage(_ context.Context, storageKey string) error {
	if p.deleteHook != nil {
		p.deleteHook()
	}
	p.deletedKeys = append(p.deletedKeys, storageKey)
	return p.deleteErr
}
