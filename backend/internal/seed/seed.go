// Package seed inserts baseline reference data (categories, tags), a bootstrap
// admin account and one worked example problem. It is idempotent: running it
// repeatedly never duplicates rows.
package seed

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/team/pkb/internal/config"
	"github.com/team/pkb/internal/shared/authx"
	"github.com/team/pkb/internal/shared/slug"
)

type categorySeed struct {
	Name        string
	Description string
	Color       string
}

var categories = []categorySeed{
	{"Backend", "API, services and server-side logic", "#2563eb"},
	{"Frontend", "UI, browser and client-side issues", "#7c3aed"},
	{"Database", "PostgreSQL, MySQL, migrations, queries", "#0891b2"},
	{"Docker", "Containers, images, compose", "#0ea5e9"},
	{"Network", "DNS, routing, firewall, connectivity", "#059669"},
	{"Wazuh", "Wazuh manager, agents, integrations", "#e11d48"},
	{"Authentication", "Login, JWT, sessions, SSO", "#d97706"},
	{"Deployment", "Release, CI/CD, rollout", "#4f46e5"},
	{"Nginx", "Reverse proxy, TLS, upstream", "#16a34a"},
	{"Infrastructure", "Servers, VMs, cloud, storage", "#64748b"},
	{"Security", "Vulnerabilities, hardening, incidents", "#dc2626"},
}

var tags = []string{
	"api", "docker", "wazuh", "network", "firewall", "nginx", "timeout",
	"database", "postgresql", "mysql", "authentication", "jwt", "deployment",
	"webhook",
}

// Run applies the full seed set.
func Run(ctx context.Context, db *gorm.DB, cfg *config.Config) error {
	if err := seedCategories(ctx, db); err != nil {
		return err
	}
	if err := seedTags(ctx, db); err != nil {
		return err
	}
	if err := seedAdmin(ctx, db, cfg); err != nil {
		return err
	}
	return seedExampleProblem(ctx, db)
}

func seedCategories(ctx context.Context, db *gorm.DB) error {
	rows := make([]map[string]any, len(categories))
	for i, c := range categories {
		rows[i] = map[string]any{
			"name": c.Name, "slug": slug.Make(c.Name),
			"description": c.Description, "color": c.Color,
		}
	}
	return db.WithContext(ctx).Table("problem_categories").
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&rows).Error
}

func seedTags(ctx context.Context, db *gorm.DB) error {
	rows := make([]map[string]any, len(tags))
	for i, name := range tags {
		rows[i] = map[string]any{"name": name, "slug": slug.Make(name)}
	}
	return db.WithContext(ctx).Table("tags").
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).
		Create(&rows).Error
}

func seedAdmin(ctx context.Context, db *gorm.DB, cfg *config.Config) error {
	var count int64
	if err := db.WithContext(ctx).Table("users").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := authx.HashPassword(cfg.SeedAdminPassword)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table("users").Create(map[string]any{
		"username":      strings.ToLower(cfg.SeedAdminUsername),
		"email":         strings.ToLower(cfg.SeedAdminEmail),
		"password_hash": hash,
		"display_name":  "Administrator",
		"role":          "ADMIN",
	}).Error
}

func seedExampleProblem(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Table("problems").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var categoryID int64
		if err := tx.Raw(`SELECT id FROM problem_categories WHERE slug = ?`, "network").
			Scan(&categoryID).Error; err != nil {
			return err
		}
		var adminID *int64
		var id int64
		if err := tx.Raw(`SELECT id FROM users ORDER BY id ASC LIMIT 1`).Scan(&id).Error; err == nil && id != 0 {
			adminID = &id
		}

		exampleTags := []string{"wazuh", "webhook", "network", "firewall", "timeout"}
		tagsCached := strings.Join(exampleTags, " ")
		solvedAt := time.Now().Add(-48 * time.Hour)

		var problemID int64
		err := tx.Raw(`
			INSERT INTO problems
				(title, description, error_message, category_id, severity, status,
				 environment, project, root_cause, solution, prevention, tags_cached,
				 created_by, updated_by, created_at, updated_at, solved_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?, now() - interval '3 days', now() - interval '2 days', ?)
			RETURNING id`,
			"Wazuh Server ส่ง Webhook แล้ว Timeout",
			"Wazuh Server ไม่สามารถส่ง Webhook ไปยัง External Webhook Server ได้ ทำให้ Alert ไม่ถูกส่งต่อไปยังระบบปลายทาง",
			"Connection to webhook-server.internal.go.th timed out\ncurl: (28) Failed to connect to webhook-server.internal.go.th port 443 after 21000 ms",
			categoryID, "HIGH", "SOLVED", "PROD", "Wazuh Backend",
			"Firewall ฝั่งปลายทางไม่ได้ Allow Source IP ของ Wazuh Server จึงทำให้ TCP handshake ไม่สำเร็จและ connection timeout",
			"ประสานทีม Network เพิ่ม Source IP ของ Wazuh Server (10.20.30.40) เข้า Firewall Policy ขา outbound ไปยัง webhook server port 443 จากนั้นทดสอบส่ง Webhook สำเร็จ",
			"จัดทำ Network Flow และ Firewall Requirement สำหรับ Wazuh Server ทุก Location และ review ทุกครั้งก่อน onboard integration ใหม่",
			tagsCached, adminID, adminID, solvedAt,
		).Scan(&problemID).Error
		if err != nil {
			return err
		}

		steps := []struct{ action, result string }{
			{"ตรวจสอบ DNS: nslookup webhook-server.internal.go.th", "DNS resolve ได้ปกติ ชี้ไปที่ 172.16.8.15"},
			{"ทดสอบ curl ไปยัง endpoint จาก Wazuh Server", "Connection timeout หลังรอ 21 วินาที"},
			{"ตรวจสอบ routing / traceroute", "packet หลุดหลัง hop ของ core switch — สงสัย firewall"},
			{"ขอทีม Network ตรวจสอบ Firewall Policy", "พบว่าไม่มี rule allow source IP ของ Wazuh Server"},
			{"เพิ่ม Firewall rule แล้วทดสอบ Webhook อีกครั้ง", "ส่ง Webhook สำเร็จ HTTP 200 ภายใน 300ms"},
		}
		for i, s := range steps {
			if err := tx.Exec(`
				INSERT INTO problem_steps (problem_id, step_no, action, result)
				VALUES (?,?,?,?)`, problemID, i+1, s.action, s.result).Error; err != nil {
				return err
			}
		}

		for _, name := range exampleTags {
			if err := tx.Exec(`
				INSERT INTO problem_tags (problem_id, tag_id)
				SELECT ?, id FROM tags WHERE name = ?
				ON CONFLICT DO NOTHING`, problemID, name).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
