package templates

import (
	"bytes"
	"fmt"
	"text/template"

	"shopvault/internal/models"
)

type OrderEmailData struct {
	ID              int64
	UserName        string
	Total           float64
	ShippingAddress string
}

func RenderReceipt(order models.Order, userName string) (string, error) {
	buf := new(bytes.Buffer)

	bodyText := fmt.Sprintf(`Order Confirmation #%d

Dear %s,

Thank you for your purchase! Your order has been received and will be shipped to:

%s

Order Total: $%.2f
Status: %s

Thank you for shopping at ShopVault!`,
		order.ID, userName, order.ShippingAddress, order.Total, order.Status,
	)

	tmpl, err := template.New("receipt").Parse(bodyText)
	if err != nil {
		return "", err
	}

	data := OrderEmailData{
		ID:              order.ID,
		UserName:        userName,
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
	}

	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
