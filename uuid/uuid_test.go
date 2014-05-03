package uuid

import (
	"testing"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestUUIDGen(c *C) {
	uuid := NewUUID()
	c.Assert(uuid, Matches, "*-*-*-*")
	c.Assert(len(uuid), Equals, 36)
}
