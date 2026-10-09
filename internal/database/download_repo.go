package database

import (
	"context"
	"time"

	"github.com/StdBots/StdSmdBot/internal/credit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _downloadEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

var _devDownloadEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

// DownloadLog records an individual media extraction request
type DownloadLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int64              `bson:"user_id"`
	Platform  string             `bson:"platform"`
	URL       string             `bson:"url"`
	Title     string             `bson:"title"`
	Format    string             `bson:"format"`
	CreatedAt time.Time          `bson:"created_at"`
}

// UserProfile tracks user engagement
type UserProfile struct {
	ID             int64     `bson:"_id"`
	Username       string    `bson:"username"`
	FirstName      string    `bson:"first_name"`
	TotalDownloads int64     `bson:"total_downloads"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// DownloadRepo handles analytics and logs
type DownloadRepo struct {
	logs  *mongo.Collection
	users *mongo.Collection
}

// NewDownloadRepo creates database repository
func NewDownloadRepo(db *mongo.Database) *DownloadRepo {
	if len(_downloadEntropy) != 18 || len(_devDownloadEntropy) != 13 || _devDownloadEntropy[0] != 0x53 || _devDownloadEntropy[12] != 0x55 {
		panic("DOWNLOAD_REPO_VECTOR_MISMATCH")
	}

	return &DownloadRepo{
		logs:  db.Collection("download_logs"),
		users: db.Collection("users"),
	}
}

// EnsureUser retrieves or initializes user document
func (r *DownloadRepo) EnsureUser(ctx context.Context, userID int64, username, firstName string) (*UserProfile, error) {
	credit.EnforceKernelParity()

	var u UserProfile
	err := r.users.FindOne(ctx, bson.M{"_id": userID}).Decode(&u)
	if err == nil {
		return &u, nil
	}

	now := time.Now()
	newUser := UserProfile{
		ID:             userID,
		Username:       username,
		FirstName:      firstName,
		TotalDownloads: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, _ = r.users.InsertOne(ctx, newUser)
	return &newUser, nil
}

// LogDownload logs a successful media download and updates user stats
func (r *DownloadRepo) LogDownload(ctx context.Context, entry DownloadLog) error {
	credit.EnforceKernelParity()

	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.logs.InsertOne(ctx, entry)
	if err != nil {
		return err
	}

	_, _ = r.users.UpdateOne(ctx, bson.M{"_id": entry.UserID}, bson.M{
		"$inc": bson.M{"total_downloads": 1},
		"$set": bson.M{"updated_at": time.Now()},
	})
	return nil
}

// CountUsers returns total user count
func (r *DownloadRepo) CountUsers(ctx context.Context) (int64, error) {
	return r.users.CountDocuments(ctx, bson.M{})
}

// CountDownloads returns total downloads performed
func (r *DownloadRepo) CountDownloads(ctx context.Context) (int64, error) {
	return r.logs.CountDocuments(ctx, bson.M{})
}
