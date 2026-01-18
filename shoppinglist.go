package appie

import (
	"context"
	"fmt"
	"net/http"
)

// listResponse matches the API response for shopping lists.
type listResponse struct {
	ID                 string     `json:"id"`
	Description        string     `json:"description"`
	ItemCount          int        `json:"itemCount"`
	HasFavoriteProduct bool       `json:"hasFavoriteProduct"`
	ProductImages      [][]string `json:"productImages"`
}

// GetShoppingLists retrieves all shopping lists.
func (c *Client) GetShoppingLists(ctx context.Context) ([]ShoppingList, error) {
	var result []listResponse
	if err := c.doRequest(ctx, http.MethodGet, "/mobile-services/lists/v3/lists", nil, &result); err != nil {
		return nil, fmt.Errorf("get shopping lists failed: %w", err)
	}

	lists := make([]ShoppingList, 0, len(result))
	for _, r := range result {
		lists = append(lists, ShoppingList{
			ID:        r.ID,
			Name:      r.Description,
			ItemCount: r.ItemCount,
		})
	}

	return lists, nil
}

// GetShoppingList retrieves the first shopping list.
// Use GetShoppingLists to get all lists if you have multiple.
func (c *Client) GetShoppingList(ctx context.Context) (*ShoppingList, error) {
	lists, err := c.GetShoppingLists(ctx)
	if err != nil {
		return nil, err
	}

	if len(lists) == 0 {
		return nil, fmt.Errorf("no shopping lists found")
	}

	return &lists[0], nil
}

// AddToShoppingList adds items to the shopping list.
// Note: The exact endpoint for adding items needs discovery.
func (c *Client) AddToShoppingList(ctx context.Context, items []ListItem) error {
	body := map[string]any{
		"items": items,
	}

	if err := c.doRequest(ctx, http.MethodPost, "/mobile-services/lists/v3/lists/items", body, nil); err != nil {
		return fmt.Errorf("add to shopping list failed: %w", err)
	}

	return nil
}

// AddProductToShoppingList adds a product to the shopping list by product ID.
func (c *Client) AddProductToShoppingList(ctx context.Context, productID int, quantity int) error {
	if quantity <= 0 {
		quantity = 1
	}

	items := []ListItem{{
		ProductID: productID,
		Quantity:  quantity,
	}}

	return c.AddToShoppingList(ctx, items)
}

// AddFreeTextToShoppingList adds a free-text item to the shopping list.
func (c *Client) AddFreeTextToShoppingList(ctx context.Context, name string, quantity int) error {
	if quantity <= 0 {
		quantity = 1
	}

	items := []ListItem{{
		Name:     name,
		Quantity: quantity,
	}}

	return c.AddToShoppingList(ctx, items)
}

// RemoveFromShoppingList removes an item from the shopping list.
func (c *Client) RemoveFromShoppingList(ctx context.Context, itemID string) error {
	path := fmt.Sprintf("/mobile-services/lists/v3/lists/items/%s", itemID)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("remove from shopping list failed: %w", err)
	}

	return nil
}

// CheckShoppingListItem marks an item as checked or unchecked.
func (c *Client) CheckShoppingListItem(ctx context.Context, itemID string, checked bool) error {
	body := map[string]any{
		"checked": checked,
	}

	path := fmt.Sprintf("/mobile-services/lists/v3/lists/items/%s", itemID)
	if err := c.doRequest(ctx, http.MethodPatch, path, body, nil); err != nil {
		return fmt.Errorf("check shopping list item failed: %w", err)
	}

	return nil
}

// ClearShoppingList removes all items from the shopping list.
func (c *Client) ClearShoppingList(ctx context.Context) error {
	list, err := c.GetShoppingList(ctx)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		if err := c.RemoveFromShoppingList(ctx, item.ID); err != nil {
			return fmt.Errorf("failed to remove item %s: %w", item.ID, err)
		}
	}

	return nil
}

// ShoppingListToOrder adds all unchecked items from the shopping list to the order.
func (c *Client) ShoppingListToOrder(ctx context.Context) error {
	list, err := c.GetShoppingList(ctx)
	if err != nil {
		return err
	}

	var orderItems []OrderItem
	for _, item := range list.Items {
		if !item.Checked && item.ProductID > 0 {
			orderItems = append(orderItems, OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
	}

	if len(orderItems) == 0 {
		return nil
	}

	return c.AddToOrder(ctx, orderItems)
}
