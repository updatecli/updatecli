module example.com/testmodule

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1 // what you'd expect to use
    github.com/sirupsen/logrus v1.8.1
)

replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0 // force older version