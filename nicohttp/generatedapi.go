package nicohttp


import (
	"fmt"
	"strings"

	"github.com/gorilla/mux"
)

func isBase(path string) bool {
	startsWith := []string{"/api", "/logs", "/dumplog", "/uptime", "/healthz", "/suspend", "/restart", "/shutdown", "/builder"}
	for _, v := range startsWith {
		if b := strings.HasPrefix(path, v); b {
			return true
		}

	}
	return false;
}


func generateAPI(httpRouter *mux.Router) ([]string, []string, error) {
	inherited := make([]string, 0)
	service := make([]string, 0)
	err := httpRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err1 := route.GetPathTemplate()
		methods, err2 := route.GetMethods()
		queriesTemplate, _ := route.GetQueriesTemplates() // bug in mux


		if (err1 == nil && err2 == nil) {
			var m, s string
			if err2 == nil {
				m = strings.Join(methods, ",")
			}
			if (len(queriesTemplate) == 0) {
				s = fmt.Sprintf("%-30s%4s%6s%4s%s", route.GetName(), "", m, " ", pathTemplate)
			} else {
				s = fmt.Sprintf("%-30s%4s%6s%4s%s?%s", route.GetName(), "", m, " ", pathTemplate, strings.Join(queriesTemplate, ","))
			}
			if (isBase(pathTemplate)) {
				inherited = append(inherited, s)
			} else {
				service = append(service, s)
			}
			return nil
		}
		return fmt.Errorf("PathTemplateError: %s, MethodsError: %s", err1, err2)
	})
	if (err == nil) {
		return inherited, service, nil
	}
	return nil, nil, err
}
