// +build windows

package client

func (s *CmdCtlStart) Run() (err error) {
	err = startService()

	return err

}
