package controller

func personalRouteScopeAnyOf(route personalRoute) [][]string {
	if len(route.capabilities) == 0 {
		return [][]string{{"media.upload"}}
	}
	result := make([][]string, 0, len(route.capabilities))
	for _, capability := range route.capabilities {
		result = append(result, []string{string(capability)})
	}
	return result
}

// PersonalAPIDeclaration exports the same PAT permission, route and media
// declarations used by runtime authorization.
func PersonalAPIDeclaration() map[string]any {
	permissions := make([]map[string]any, 0, len(personalPermissions))
	for _, permission := range personalPermissions {
		permissions = append(permissions, map[string]any{
			"key":         permission.Key,
			"label":       permission.Label,
			"description": permission.Description,
			"capability":  string(permission.Capability),
		})
	}

	routes := make([]map[string]any, 0, len(personalRoutes)+len(personalPublicReadPaths))
	for _, path := range personalPublicReadPaths {
		routes = append(routes, map[string]any{
			"method": "GET", "path": path,
			"scopeAnyOf": [][]string{{}}, "authentication": "public-or-pat",
		})
	}
	for _, route := range personalRoutes {
		routes = append(routes, map[string]any{
			"method": route.method, "path": route.path,
			"scopeAnyOf": personalRouteScopeAnyOf(route), "authentication": "pat",
		})
	}

	media := make([]map[string]any, 0, len(personalMediaProfiles))
	for _, profile := range personalMediaProfiles {
		media = append(media, map[string]any{
			"namespace": "gallery", "profile": profile.profile,
			"scopeAnyOf": [][]string{{profile.permission}},
		})
	}

	return map[string]any{
		"schemaVersion": 1,
		"product": map[string]string{
			"id": "gallery", "name": "Gallery",
			"guide": "../developer-tokens.md", "openapi": "../../contracts/openapi/gallery.json",
		},
		"permissions": permissions, "routes": routes, "mediaProfiles": media,
	}
}
