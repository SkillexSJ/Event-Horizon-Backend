package store

import (
	"context"
	"errors"
	"event-horizon/models"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/******************** MONGODB FUNCTIONALITY FOR CATEGORY COLLECTIONS *********************

1. BSON MAPPING FOR CATEGORY COLLECTION
2. InsertOne
3. FindOne
4. Find
5. DeleteOne
6. UpdateOne
7. CountDocuments
8. Regex Search
9. Case-Insensitive Search

 ****************************************************************************************/

/************************** I DID THE FOLLOWING THINGS IN THIS FILE: *******************************************


1. Created CategoryStore struct to manage category-related database operations.

2. Implemented NewCategoryStore constructor to initialize CategoryStore with MongoDB collections.

3. Developed CreateCategory method to add new categories, ensuring no duplicates by name.

4. Created GetAllCategories method to retrieve all categories from the database.

5. Added GetAllCategoriesWithEvents method to fetch categories along with their associated events.

6. Implemented GetCategoryByID method to fetch a specific category by its ID.

7. Developed GetCategoryByName method to find a category by name using case-insensitive search.

8. Created GetCategoryWithEvents method to retrieve a category along with all its events.

9. Added helper method getEventsByCategory to fetch events for a given category.

10. Implemented GetCategoryEventCount method to count events in a category.

11. Developed DeleteCategory method to remove a category only if it has no associated events.

12. Created UpdateCategory method to modify existing category details.


************************************************************************************************************/

type CategoryStore struct {
	collection      *mongo.Collection
	eventCollection *mongo.Collection
	bookingStore    *BookingStore
}

func NewCategoryStore(db *mongo.Database) *CategoryStore {
	return &CategoryStore{
		collection:      db.Collection("Categories"),
		eventCollection: db.Collection("Events"),
		bookingStore:    nil, // Will be set later
	}
}

// SetBookingStore sets the bookingStore reference for cascade deletes
func (s *CategoryStore) SetBookingStore(bookingStore *BookingStore) {
	s.bookingStore = bookingStore
}

func (s *CategoryStore) CreateCategory(ctx context.Context, category *models.Category) error {

	filter := bson.M{"name": category.Name} //! Exact match filter
	var existingCategory models.Category

	err := s.collection.FindOne(ctx, filter).Decode(&existingCategory)
	if err == nil {
		return errors.New("category with the same name already exists")
	}

	//? if the error is not ErrNoDocuments, return the error
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	//? Set creation timestamp
	category.CreatedAt = time.Now()

	result, err := s.collection.InsertOne(ctx, category)
	if err != nil {
		return err
	}

	category.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetAllCategories retrieves all categories
func (s *CategoryStore) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []models.Category{}
	}

	return categories, nil
}

// GetAllCategoriesWithEvents retrieves all categories with their event counts
func (s *CategoryStore) GetAllCategoriesWithEvents(ctx context.Context) ([]models.CategoryWithEvents, error) {
	categories, err := s.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	var result []models.CategoryWithEvents
	for _, category := range categories {
		//? Get events for this category
		events, err := s.getEventsByCategory(ctx, category.ID)
		if err != nil {
			events = []models.Event{} // Empty on error
		}

		result = append(result, models.CategoryWithEvents{
			Category:   category,
			Events:     events,
			EventCount: len(events),
		})
	}

	return result, nil
}

// GetCategoryByID retrieves a category by ID
func (s *CategoryStore) GetCategoryByID(ctx context.Context, categoryID bson.ObjectID) (*models.Category, error) {
	var category models.Category

	err := s.collection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			println("DEBUG Store: No documents found for ID")
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// GetCategoryByName retrieves a category by name (case-insensitive)
func (s *CategoryStore) GetCategoryByName(ctx context.Context, categoryName string) (*models.Category, error) {
	var category models.Category

	escapedName := regexp.QuoteMeta(categoryName)

	//? Use case-insensitive regex to match category name
	filter := bson.M{
		"name": bson.M{
			"$regex":   "^" + escapedName + "$",
			"$options": "i", //! case-insensitive
		},
	}

	err := s.collection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			println("DEBUG Store: No documents found")
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// GetCategoryWithEvents retrieves a category with all its events
func (s *CategoryStore) GetCategoryWithEvents(ctx context.Context, categoryID bson.ObjectID) (*models.CategoryWithEvents, error) {
	//? Get category
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	//? Get all events under this category
	events, err := s.getEventsByCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	return &models.CategoryWithEvents{
		Category:   *category,
		Events:     events,
		EventCount: len(events),
	}, nil
}

// getEventsByCategory is a helper method to get all events for a category
func (s *CategoryStore) getEventsByCategory(ctx context.Context, categoryID bson.ObjectID) ([]models.Event, error) {
	var events []models.Event

	//? First get the category to find its name
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	//? Search events by category_name instead of category_id
	filter := bson.M{"category_name": category.Name} //! Match by NAME
	cursor, err := s.eventCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	if events == nil {
		events = []models.Event{}
	}

	return events, nil
}

///// PORE KAJ HOBE /////

// GetCategoryEventCount returns the number of events in a category
func (s *CategoryStore) GetCategoryEventCount(ctx context.Context, categoryID bson.ObjectID) (int, error) {
	//? Get category to find its name
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return 0, err
	}

	//? Count events by category_name instead of category_id
	count, err := s.eventCollection.CountDocuments(ctx, bson.M{"category_name": category.Name})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// DeleteCategory deletes a category (only if it has no events)
func (s *CategoryStore) DeleteCategory(ctx context.Context, categoryID bson.ObjectID) error {
	//? Check if category has any events
	count, err := s.GetCategoryEventCount(ctx, categoryID)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete category with existing events")
	}

	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": categoryID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// DeleteCategoryWithCascade deletes a category and all its associated events and bookings
func (s *CategoryStore) DeleteCategoryWithCascade(ctx context.Context, categoryID bson.ObjectID) error {
	//? Get category first to get its name
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return err
	}

	//? Get all events under this category to delete their bookings
	events, err := s.getEventsByCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	//? Delete bookings for each event first (if bookingStore is available)
	if s.bookingStore != nil {
		for _, event := range events {
			_, err := s.bookingStore.DeleteBookingsByEventID(ctx, event.ID)
			if err != nil {
				return errors.New("failed to delete bookings for event: " + err.Error())
			}
		}
	}

	//? Delete all events by category_name
	filter := bson.M{"category_name": category.Name}
	_, err = s.eventCollection.DeleteMany(ctx, filter)
	if err != nil {
		return errors.New("failed to delete events: " + err.Error())
	}

	//? Now delete the category itself
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": categoryID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// UpdateCategory updates a category's details
func (s *CategoryStore) UpdateCategory(ctx context.Context, categoryID bson.ObjectID, updates bson.M) error {
	filter := bson.M{"_id": categoryID} //! Match by ID
	update := bson.M{"$set": updates}   //! Set the updates

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}
