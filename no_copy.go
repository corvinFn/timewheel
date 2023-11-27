package timewheel

type noCopy struct{} //nolint: unused

func (*noCopy) Lock()   {} //nolint: unused
func (*noCopy) Unlock() {} //nolint: unused
