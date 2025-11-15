package paprika_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/soggycactus/paprika-3-mcp/internal/paprika"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	testRecipe := paprika.Recipe{
		Name:        fmt.Sprintf("Test Recipe - %d", time.Now().Unix()),
		Notes:       "Notes",
		Directions:  "Directions",
		Ingredients: "Ingredients",
		Servings:    "Servings",
		Source:      "Source",
		SourceURL:   "URL",
		Categories:  []string{},
	}
	recipe, err := client.SaveRecipe(ctx, testRecipe)
	require.NoError(t, err)

	recipe, err = client.GetRecipe(ctx, recipe.UID)
	require.NoError(t, err)
	assert.NotEmpty(t, recipe.UID)
	assert.Equal(t, testRecipe.Name, recipe.Name)
	assert.Equal(t, testRecipe.Notes, recipe.Notes)
	assert.Equal(t, testRecipe.Directions, recipe.Directions)
	assert.Equal(t, testRecipe.Ingredients, recipe.Ingredients)
	assert.Equal(t, testRecipe.Servings, recipe.Servings)
	assert.Equal(t, testRecipe.Source, recipe.Source)
	assert.Equal(t, testRecipe.SourceURL, recipe.SourceURL)
	assert.Equal(t, testRecipe.Categories, recipe.Categories)

	t.Logf("Created and fetched recipe: %+v", recipe)

	newDescription := "Updated Description"
	recipe.Description = newDescription
	uid := recipe.UID
	recipe, err = client.SaveRecipe(ctx, *recipe)
	require.NoError(t, err)
	assert.Equal(t, newDescription, recipe.Description)
	assert.Equal(t, uid, recipe.UID)
	assert.Equal(t, testRecipe.Name, recipe.Name)
	assert.Equal(t, testRecipe.Notes, recipe.Notes)
	assert.Equal(t, testRecipe.Directions, recipe.Directions)
	assert.Equal(t, testRecipe.Ingredients, recipe.Ingredients)
	assert.Equal(t, testRecipe.Servings, recipe.Servings)
	assert.Equal(t, testRecipe.Source, recipe.Source)
	assert.Equal(t, testRecipe.SourceURL, recipe.SourceURL)
	assert.Equal(t, testRecipe.Categories, recipe.Categories)

	t.Logf("Updated recipe: %+v", recipe)

	_, err = client.DeleteRecipe(ctx, *recipe)
	require.NoError(t, err)
	t.Logf("Deleted recipe: %s", recipe.Name)

	recipes, err := client.ListRecipes(ctx)
	require.NoError(t, err)

	for _, recipe := range recipes.Result {
		r, err := client.GetRecipe(ctx, recipe.UID)
		require.NoError(t, err)

		t.Logf("Recipe: %s - %s", r.Name, r.Created)
		if _, err := json.Marshal(r); err != nil {
			t.Logf("Failed to marshal recipe: %s", err)
		}
	}
}

func TestMealPlan(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test listing meal plan (should work even if empty)
	mealPlanResp, err := client.ListMealPlan(ctx)
	require.NoError(t, err)
	assert.NotNil(t, mealPlanResp)
	t.Logf("Found %d existing meals in plan", len(mealPlanResp.Result))

	// Create a test meal plan entry
	testDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02") // 7 days from now
	testMeal := paprika.MealPlan{
		Date:      testDate + " 00:00:00",
		Name:      fmt.Sprintf("Test Meal - %d", time.Now().Unix()),
		Type:      2, // Dinner
		RecipeUID: "",
		OrderFlag: 0,
	}

	savedMeal, err := client.SaveMealPlan(ctx, testMeal)
	require.NoError(t, err)
	assert.NotEmpty(t, savedMeal.UID)
	assert.Equal(t, testMeal.Name, savedMeal.Name)
	assert.Equal(t, testMeal.Type, savedMeal.Type)
	t.Logf("Created meal plan entry: %+v", savedMeal)

	// Update the meal plan entry
	savedMeal.Name = "Updated Test Meal"
	savedMeal.Type = 0 // Change to Breakfast
	updatedMeal, err := client.SaveMealPlan(ctx, *savedMeal)
	require.NoError(t, err)
	assert.Equal(t, savedMeal.UID, updatedMeal.UID)
	assert.Equal(t, "Updated Test Meal", updatedMeal.Name)
	assert.Equal(t, 0, updatedMeal.Type)
	t.Logf("Updated meal plan entry: %+v", updatedMeal)

	// List meal plan again to verify it's there
	mealPlanResp, err = client.ListMealPlan(ctx)
	require.NoError(t, err)
	found := false
	for _, meal := range mealPlanResp.Result {
		if meal.UID == savedMeal.UID {
			found = true
			assert.Equal(t, "Updated Test Meal", meal.Name)
			assert.Equal(t, 0, meal.Type)
			break
		}
	}
	assert.True(t, found, "Updated meal should be in the meal plan list")

	// Delete the meal plan entry
	err = client.DeleteMealPlan(ctx, savedMeal.UID)
	require.NoError(t, err)
	t.Logf("Deleted meal plan entry: %s", savedMeal.UID)

	// Verify deletion
	mealPlanResp, err = client.ListMealPlan(ctx)
	require.NoError(t, err)
	for _, meal := range mealPlanResp.Result {
		assert.NotEqual(t, savedMeal.UID, meal.UID, "Deleted meal should not be in the list")
	}
}

func TestGroceries(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test listing groceries (should work even if empty)
	groceryResp, err := client.ListGroceries(ctx)
	require.NoError(t, err)
	assert.NotNil(t, groceryResp)
	t.Logf("Found %d existing grocery items", len(groceryResp.Result))

	// Create a test grocery item
	testItem := paprika.GroceryItem{
		Ingredient:  fmt.Sprintf("Test Item - %d", time.Now().Unix()),
		Name:        fmt.Sprintf("Test Item - %d", time.Now().Unix()),
		Aisle:       "Test Aisle",
		Quantity:    "2",
		Recipe:      "Test Recipe",
		Instruction: "Test instruction",
		Purchased:   false,
		OrderFlag:   0,
	}

	savedItem, err := client.SaveGroceryItem(ctx, testItem)
	require.NoError(t, err)
	assert.NotEmpty(t, savedItem.UID)
	assert.Equal(t, testItem.Ingredient, savedItem.Ingredient)
	assert.Equal(t, testItem.Aisle, savedItem.Aisle)
	assert.Equal(t, testItem.Quantity, savedItem.Quantity)
	assert.False(t, savedItem.Purchased)
	t.Logf("Created grocery item: %+v", savedItem)

	// Update the grocery item
	savedItem.Ingredient = "Updated Test Item"
	savedItem.Quantity = "5"
	savedItem.Purchased = true
	updatedItem, err := client.SaveGroceryItem(ctx, *savedItem)
	require.NoError(t, err)
	assert.Equal(t, savedItem.UID, updatedItem.UID)
	assert.Equal(t, "Updated Test Item", updatedItem.Ingredient)
	assert.Equal(t, "5", updatedItem.Quantity)
	assert.True(t, updatedItem.Purchased)
	t.Logf("Updated grocery item: %+v", updatedItem)

	// List groceries again to verify it's there
	groceryResp, err = client.ListGroceries(ctx)
	require.NoError(t, err)
	found := false
	for _, item := range groceryResp.Result {
		if item.UID == savedItem.UID {
			found = true
			assert.Equal(t, "Updated Test Item", item.Ingredient)
			assert.Equal(t, "5", item.Quantity)
			assert.True(t, item.Purchased)
			break
		}
	}
	assert.True(t, found, "Updated grocery item should be in the list")

	// Delete the grocery item
	err = client.DeleteGroceryItem(ctx, savedItem.UID)
	require.NoError(t, err)
	t.Logf("Deleted grocery item: %s", savedItem.UID)

	// Verify deletion
	groceryResp, err = client.ListGroceries(ctx)
	require.NoError(t, err)
	for _, item := range groceryResp.Result {
		assert.NotEqual(t, savedItem.UID, item.UID, "Deleted item should not be in the list")
	}
}
