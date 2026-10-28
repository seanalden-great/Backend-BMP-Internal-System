// controllers/auth_controller.go
package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"accounting-be/models"
)

type AuthController struct {
	DB *gorm.DB
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY")) // Pastikan Anda menambahkan variabel ini di Vercel nantinya

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	FullName    string `json:"full_name"`
	Role        string `json:"role"`
	Permissions string `json:"permissions"`
}

// POST /api/auth/register (Biasanya hanya boleh diakses oleh Admin)
func (ac *AuthController) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid"})
		return
	}

	// Cek apakah username sudah dipakai
	var existingUser models.User
	if err := ac.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username sudah terdaftar"})
		return
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	user := models.User{
		Username:    input.Username,
		Password:    string(hashedPassword),
		FullName:    input.FullName,
		Role:        input.Role,
		Permissions: input.Permissions,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user baru"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User berhasil didaftarkan!"})
}

// POST /api/auth/login
func (ac *AuthController) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan Password wajib diisi"})
		return
	}

	// 1. Cari user di database
	var user models.User
	if err := ac.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah"})
		return
	}

	// 2. Bandingkan password asli dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah"})
		return
	}

	// 3. Buat JWT Token
	// Kalau JWT_SECRET belum diset (misal di lokal), pakai default sementara
	secret := jwtSecret
	if len(secret) == 0 {
		secret = []byte("rahasia_perusahaan_bmp_123!") 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"permissions": user.Permissions,
		"exp":         time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 Jam
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meng-generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  tokenString,
		"user": map[string]interface{}{
			"id":          user.ID,
			"username":    user.Username,
			"full_name":   user.FullName,
			"role":        user.Role,
			"permissions": user.Permissions,
		},
	})
}