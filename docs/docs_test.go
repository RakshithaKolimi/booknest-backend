package docs

import "testing"

func TestSwaggerInfoInitialized(t *testing.T) {
	if SwaggerInfo.Title == "" {
		t.Fatal("expected swagger title to be initialized")
	}
	if SwaggerInfo.Version == "" {
		t.Fatal("expected swagger version to be initialized")
	}
	if SwaggerInfo.BasePath == "" {
		t.Fatal("expected swagger base path to be initialized")
	}
	if SwaggerInfo.Schemes == nil || len(SwaggerInfo.Schemes) == 0 {
		t.Fatal("expected swagger schemes to be initialized")
	}
}
