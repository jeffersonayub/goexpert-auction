package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreateAuction_Success(t *testing.T) {
	// Set auction interval to a short duration for the test
	os.Setenv("AUCTION_INTERVAL", "10ms")
	defer os.Unsetenv("AUCTION_INTERVAL")

	mtestDB := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mtestDB.Run("CreateAuction", func(mt *mtest.T) {

		repo := NewAuctionRepository(mt.DB)

		auctionEntity := &auction_entity.Auction{
			Id:          "auction1",
			ProductName: "Test Product",
			Category:    "Electronics",
			Description: "A test product",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Unix(time.Now().Unix(), 0),
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		err := repo.CreateAuction(context.Background(), auctionEntity)
		assert.Nil(t, err)
		assert.Equal(t, "auction1", auctionEntity.Id)
		assert.Equal(t, auction_entity.Active, auctionEntity.Status)

		insertEvent := mt.GetAllStartedEvents()
		assert.Equal(t, 1, len(insertEvent))
		assert.Equal(t, "insert", insertEvent[0].CommandName)

		// Wait for a short duration to allow the goroutine to execute UpdateOne
		time.Sleep(20 * time.Millisecond)

		startedEvents := mt.GetAllStartedEvents()
		assert.Equal(t, "update", startedEvents[1].CommandName)

		var updatedAuction bson.M
		errDoc := bson.Unmarshal(startedEvents[1].Command, &updatedAuction)
		assert.Nil(t, errDoc)
		assert.Equal(t, "auctions", updatedAuction["update"])
		assert.NotEmpty(t, updatedAuction["updates"])

		// Verifica se o status foi atualizado para Completed
		updatedStatus := updatedAuction["updates"].(bson.A)[0].(bson.M)["u"].(bson.M)["$set"].(bson.M)["status"]
		assert.Equal(t, int32(auction_entity.Completed), updatedStatus)
	})
}
