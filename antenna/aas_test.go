// Implements a simple phase delay from different n-Antenna elements
package antenna_test

import (
    "testing"
    "wiless/cellular/antenna"
  
)

/// Drops Linear Vertical Nodes spaced with dh,dv linearly
func  TestNewAAS(t *testing.T)  {
	output:=antenna.NewAAS()
	t.Log("Found antenna",output)
}
