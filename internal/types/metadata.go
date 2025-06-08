package types

// MetadataOutput represents the response structure for the root endpoint ("/").
type MetadataOutput struct {
	Body struct {
		Service       string          `json:"service" example:"My API"`
		Version       string          `json:"version" example:"v1"`
		Description   string          `json:"description" example:"API description"`
		Status        string          `json:"status" example:"operational"`
		Uptime        string          `json:"uptime" example:"8d 19h 16m"`
		Health        HealthStatus    `json:"health"`
		Documentation string          `json:"documentation" example:"/docs"`
		Links         MetadataLinks   `json:"links"`
		Contact       MetadataContact `json:"contact"`
		Environment   string          `json:"environment" example:"development"`
	}
}

// HealthStatus holds health check details.
type HealthStatus struct {
	Database string  `json:"database" example:"ok"`
	Server   string  `json:"server" example:"ok"`
	Load     float64 `json:"load" example:"11.35"` // Simulated load value
}

// MetadataLinks holds related links for the metadata endpoint.
type MetadataLinks struct {
	Self          string `json:"self" example:"/"`
	PrivacyPolicy string `json:"privacyPolicy" example:"/api/terms_condition"`
}

// MetadataContact holds contact information for the metadata endpoint.
type MetadataContact struct {
	Name  string `json:"name" example:"API Support"`
	Email string `json:"email" example:"mail@achyutkoirala.com.np"`
	URL   string `json:"url" example:"/contact"`
}

// HealthCheckOutput represents the response structure for the health check endpoint ("/healthz").
type HealthCheckOutput struct {
	Body struct {
		Status string `json:"status" example:"healthy" doc:"API health status"`
	}
}

// Placeholder for other types (e.g., User, Good, Transaction)
// type User struct {
//     ID   string `json:"id"`
//     Name string `json:"name"`
// }
