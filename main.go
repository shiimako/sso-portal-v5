package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"sso-portal-v5/config"
	"sso-portal-v5/controllers/admincontroller"
	"sso-portal-v5/controllers/authcontroller"
	"sso-portal-v5/controllers/dashboardcontroller"
	"sso-portal-v5/controllers/redirectcontroller"
	"sso-portal-v5/controllers/usercontroller"
	"sso-portal-v5/middleware"
	"sso-portal-v5/views"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	// Init logger lumberjack
	config.InitLogger()
	log.Println("Logger berjalan : ", time.Now())

	// Load environment variables dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Gagal memuat file .env")
	}

	// Register tipe data kustom untuk session
	gob.Register([]map[string]interface{}{}) // Untuk menyimpan slice of map di session

	// Inisialisasi database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Gagal inisialisasi database: %v", err)
	}
	defer db.Close()

	// Inisialisasi session store
	sessionStore := config.InitSessionStore()

	// Inisialisasi template renderer
	templates, err := views.InitTemplates()
	if err != nil {
		log.Fatalf("Gagal inisialisasi template: %v", err)
	}

	// Inisialisasi environment untuk handler
	env := config.NewEnv(db, sessionStore, templates)

	viewEngine := views.NewViews(env)

	// Inisialisasi Google OAuth2 config
	oauthConfig := config.InitGoogleOAuthConfig(env.BaseURL)
	env.GoogleOAuthConfig = oauthConfig

	// Load RSA keys untuk JWT
	if err := config.LoadKeys(); err != nil {
    log.Fatalf("Gagal memuat RSA keys: %v", err)
	}

	// Inisialisasi controller
	authCtrl := authcontroller.NewAuthController(env, viewEngine)
	dashboardCtrl := dashboardcontroller.NewDashboardController(env, viewEngine)
	adminCtrl := admincontroller.NewAdminController(env, viewEngine)
	redirectCtrl := redirectcontroller.NewRedirectController(env, viewEngine)
	userCtrl := usercontroller.NewUserController(env, viewEngine)

	// Setup Router
	r := mux.NewRouter()

	uploadsFs := http.FileServer(http.Dir("./public/uploads"))
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", uploadsFs),
	)
	staticFs := http.FileServer(http.Dir("./public/static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	r.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/javascript")
        http.ServeFile(w, r, "./public/sw.js")
    })

	// ===================================
	// AUTHENTICATION ROUTES
	// ====================================
	r.HandleFunc("/", authCtrl.ShowLoginPage).Methods("GET")
	r.HandleFunc("/auth/google/login", authCtrl.LoginWithGoogle).Methods("GET")
	r.HandleFunc("/auth/google/callback", authCtrl.GoogleCallback).Methods("GET")
	r.HandleFunc("/logout", authCtrl.Logout).Methods("GET")

	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.GlobalAuthMiddleware(env))
	// ===================================
	// DASHBOARD ROUTES
	// ====================================
	protected.HandleFunc("/dashboard", dashboardCtrl.Index).Methods("GET")

	// ===================================
	// USER ROUTES
	// ====================================
	protected.HandleFunc("/profile/edit", userCtrl.ShowProfileForm).Methods("GET")
	protected.HandleFunc("/profile/update", userCtrl.HandleProfileUpdate).Methods("POST")
	r.HandleFunc("/avatar/{userID}", userCtrl.ServeAvatar).Methods("GET")

	// ===================================
	// REDIRECT MANAGEMENT
	// ====================================
	protected.HandleFunc("/redirect", redirectCtrl.RedirectToApp).Methods("GET")

	// ===================================
	// ADMIN ROUTES
	// ====================================
	adminRouter := protected.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminMiddleware(env, viewEngine))
	adminRouter.HandleFunc("/dashboard", adminCtrl.Dashboard).Methods("GET")

	// ===================================
	// USER MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/users", adminCtrl.ListUsers).Methods("GET")
	adminRouter.HandleFunc("/user/detail/{id}", adminCtrl.DetailUser).Methods("GET")
	adminRouter.HandleFunc("/user/edit/{id}", adminCtrl.EditUserForm).Methods("GET")
	adminRouter.HandleFunc("/user/update/{id}", adminCtrl.UpdateUser).Methods("POST")
	adminRouter.HandleFunc("/user/delete/{id}", adminCtrl.DeleteUser).Methods("POST")
	adminRouter.HandleFunc("/user/new", adminCtrl.NewUserForm).Methods("GET")
	adminRouter.HandleFunc("/user/create", adminCtrl.CreateUser).Methods("POST")

	// ===================================
	// APPLICATION MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/applications", adminCtrl.ListApplications).Methods("GET")
	adminRouter.HandleFunc("/application/detail/{id}", adminCtrl.DetailApplication).Methods("GET")
	adminRouter.HandleFunc("/application/new", adminCtrl.NewApplicationForm).Methods("GET")
	adminRouter.HandleFunc("/application/create", adminCtrl.CreateApplication).Methods("POST")
	adminRouter.HandleFunc("/application/edit/{id}", adminCtrl.EditApplicationForm).Methods("GET")
	adminRouter.HandleFunc("/application/update/{id}", adminCtrl.UpdateApplication).Methods("POST")
	adminRouter.HandleFunc("/application/delete/{id}", adminCtrl.DeleteApplication).Methods("POST")

	// ===================================
	// MAJOR MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/majors", adminCtrl.ListMajors).Methods("GET")
	adminRouter.HandleFunc("/major/new", adminCtrl.NewMajorForm).Methods("GET")
	adminRouter.HandleFunc("/major/create", adminCtrl.CreateMajor).Methods("POST")
	adminRouter.HandleFunc("/major/edit/{id}", adminCtrl.EditMajorForm).Methods("GET")
	adminRouter.HandleFunc("/major/update/{id}", adminCtrl.UpdateMajor).Methods("POST")
	adminRouter.HandleFunc("/major/delete/{id}", adminCtrl.DeleteMajor).Methods("POST")

	// ===================================
	// STUDY PROGRAM MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/study-programs", adminCtrl.ListStudyPrograms).Methods("GET")
	adminRouter.HandleFunc("/study-program/new", adminCtrl.NewStudyProgramForm).Methods("GET")
	adminRouter.HandleFunc("/study-program/create", adminCtrl.CreateStudyProgram).Methods("POST")
	adminRouter.HandleFunc("/study-program/edit/{id}", adminCtrl.EditStudyProgramForm).Methods("GET")
	adminRouter.HandleFunc("/study-program/update/{id}", adminCtrl.UpdateStudyProgram).Methods("POST")
	adminRouter.HandleFunc("/study-program/delete/{id}", adminCtrl.DeleteStudyProgram).Methods("POST")

	// ===================================
	// POSITION MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/positions", adminCtrl.ListPositions).Methods("GET")
	adminRouter.HandleFunc("/position/new", adminCtrl.NewPositionForm).Methods("GET")
	adminRouter.HandleFunc("/position/create", adminCtrl.CreatePosition).Methods("POST")
	adminRouter.HandleFunc("/position/edit/{id}", adminCtrl.EditPositionForm).Methods("GET")
	adminRouter.HandleFunc("/position/update/{id}", adminCtrl.UpdatePosition).Methods("POST")
	adminRouter.HandleFunc("/position/delete/{id}", adminCtrl.DeletePosition).Methods("POST")

	// ===================================
	// ROLES MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/roles", adminCtrl.ListRoles).Methods("GET")
	adminRouter.HandleFunc("/role/new", adminCtrl.NewRoleForm).Methods("GET")
	adminRouter.HandleFunc("/role/create", adminCtrl.CreateRole).Methods("POST")
	adminRouter.HandleFunc("/role/edit/{id}", adminCtrl.EditRoleForm).Methods("GET")
	adminRouter.HandleFunc("/role/update/{id}", adminCtrl.UpdateRole).Methods("POST")
	adminRouter.HandleFunc("/role/delete/{id}", adminCtrl.DeleteRole).Methods("POST")

	// ===================================
	// CATEGORIES MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/categories", adminCtrl.ListCategories).Methods("GET")
	adminRouter.HandleFunc("/category/new", adminCtrl.NewCategoriesForm).Methods("GET")
	adminRouter.HandleFunc("/category/create", adminCtrl.CreateCategory).Methods("POST")
	adminRouter.HandleFunc("/category/edit/{id}", adminCtrl.EditCategoriesForm).Methods("GET")
	adminRouter.HandleFunc("/category/update/{id}", adminCtrl.UpdateCategory).Methods("POST")
	adminRouter.HandleFunc("/category/delete/{id}", adminCtrl.DeleteCategory).Methods("POST")

	// ===================================
	// API ROUTES
	// ====================================
	r.HandleFunc("/api/webhook", adminCtrl.HandleWebhook).Methods("POST")
	protected.HandleFunc("/api/push/subscribe", adminCtrl.SubscribePush).Methods("POST")

	port := os.Getenv("PORT")
	log.Printf("Server berjalan di http://localhost:%s", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
