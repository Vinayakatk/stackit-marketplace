package apis

import (
	"github.com/Vinayakatk/marketplace-prototype/internal/services/billing"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/catalog"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/projects"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/users"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux) {
	// User routes
	r.Route("/api/users", func(r chi.Router) {
		r.Post("/new", users.CreateUser) // Create a new user
		r.Get("/", users.ListUsers)      // List all users

		r.Get("/{id}/deployments", deployments.ListUserDeployments)
	})

	// Project routes
	r.Route("/api/user/project", func(r chi.Router) {
		r.Post("/new", projects.CreateProject)                        // Create a new project
		r.Get("/{id}", projects.ListProjects)                         // List projects of a user
		r.Get("/{id}/deployments", projects.GetDeploymentsOfAProject) // Get deployments of a project
		r.Delete("/{id}", projects.DeleteProject)                     // Delete project
	})

	// Application catalog routes
	r.Route("/api/apps", func(r chi.Router) {
		r.Post("/new", catalog.AddApplication)       // Add a new application
		r.Get("/", catalog.ListApplications)         // List all applications
		r.Get("/{id}", catalog.GetApplication)       // Get app details
		r.Put("/{id}", catalog.UpdateApplication)    // Update app
		r.Delete("/{id}", catalog.DeleteApplication) // Delete app
	})

	// Deployment routes
	r.Route("/api/deployments", func(r chi.Router) {
		r.Post("/install", deployments.DeployApplication) // Deploy an application

		r.Get("/{id}", deployments.GetDeployment)       // Get deployment details
		r.Delete("/{id}", deployments.DeleteDeployment) // Remove deployment
	})

	// Billing apis
	r.Route("/api/billing", func(r chi.Router) {
		r.Get("/user/{consumerID}/deployment/{deploymentID}", billing.GetBillingByUserAndDeployment)
		r.Get("/user/{id}", billing.GetUserBilling)
	})
}
