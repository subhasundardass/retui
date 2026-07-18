package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

func BasicInputExample(props retui.Props) retui.Element {

	totalFields := 9

	name, setName := retui.UseState("")
	age, setAge := retui.UseState(0)
	price, setPrice := retui.UseState(0.00)
	fromDate, setFromDate := retui.UseState("")
	toDate, setToDate := retui.UseState("")
	password, setPassword := retui.UseState("")
	selectPicker, setSelectPicker := retui.UseState(0)
	agree, setAgree := retui.UseState(false)

	// Focus management
	focusIndex, setFocusIndex := retui.UseState(0)
	contentFocused := retui.IsFocused("content")

	if retui.IsFocused("content") {
		switch retui.CurrentKey.Code {
		case retui.KeyEscape:
			retui.SetFocus("sidebar")
		case retui.KeyDown:
			setFocusIndex((focusIndex + 1) % totalFields)
		case retui.KeyUp:
			setFocusIndex((focusIndex - 1 + totalFields) % totalFields)
		}
	}
	isFocused := func(idx int) bool { return contentFocused && focusIndex == idx }

	return retui.Box(
		props,
		retui.NewStyle(),
		components.Panel().
			Header(retui.Text("Basic Inputs", retui.NewStyle())).
			Width(retui.Fixed(100)).
			Children(

				// Name
				retui.Box(
					retui.Props{Gap: 1, Width: retui.Grow(1)},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("Name", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.TextInput().
							ID("name").
							Focused(isFocused(0)).
							Value(name).
							Placeholder("Enter Name").
							Style(retui.NewStyle().Bold(true)).
							OnChange(func(id string, value string) {
								setName(value)
							}).
							Render(),
					),
				),

				// Age
				retui.Box(
					retui.Props{Gap: 1, Width: retui.Grow(1)},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("Age", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.NumberInput().
							ID("age").
							Focused(isFocused(1)).
							Value(float64(age)).
							Placeholder("Enter age").
							Min(1).
							Max(100).
							ArrowStep(false).
							OnChange(func(id string, value float64) {
								setAge(int(value))
							}).
							Render(),
					),
				),

				// Price
				retui.Box(
					retui.Props{Gap: 1, Width: retui.Grow(1)},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("Price", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.NumberInput().
							ID("price").
							Focused(isFocused(2)).
							Value(price).
							Placeholder("Price").
							Decimals(2).
							ArrowStep(false).
							OnChange(func(id string, value float64) {
								setPrice(value)
							}).
							Render(),
					),
				),

				// From Date
				retui.Box(
					retui.Props{Gap: 1},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("From Date", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.DateInput().
							ID("fromDate").
							Focused(isFocused(3)).
							Value(fromDate).
							Format("DD/MM/YYYY").
							OnChange(func(id string, value string) {
								setFromDate(value)
							}).
							Render(),
					),
				),

				// To Date
				retui.Box(
					retui.Props{Gap: 1},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("To Date", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.DateInput().
							ID("toDate").
							Focused(isFocused(4)).
							Value(toDate).
							Format("DD/MM/YYYY").
							OnChange(func(id string, value string) {
								setToDate(value)
							}).
							Render(),
					),
				),

				// Password
				retui.Box(
					retui.Props{Gap: 1},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("Password", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.Password().
							ID("password").
							Focused(isFocused(5)).
							Value(password).
							OnChange(func(id string, value string) {
								setPassword(value)
							}).
							Render(),
					),
				),

				// Select Picker
				retui.Box(
					retui.Props{Gap: 1},
					retui.NewStyle(),
					retui.Box(
						retui.Props{Width: retui.Fixed(10)},
						retui.NewStyle(),
						retui.Text("Select", retui.NewStyle()),
					),
					retui.Box(
						retui.Props{Width: retui.Fixed(30)},
						retui.NewStyle(),
						components.SelectPicker().
							ID("select_picker").
							Focused(isFocused(6)).
							Options([]string{"Red", "Green", "Blue", "Yellow"}).
							Selected(selectPicker).
							OnChange(func(id string, selected int, value string) {
								setSelectPicker(selected)
							}).
							Render(),
					),
				),

				// Checkbox
				retui.Box(
					retui.Props{Gap: 1},
					retui.NewStyle(),
					components.Checkbox().
						ID("agree").
						Focused(isFocused(7)).
						Checked(agree).
						Label("I agree to the terms").
						OnChange(func(id string, checked bool) {
							setAgree(checked)
						}).
						Render(),
				),

				// Submit Button
				retui.Box(
					retui.Props{Padding: [4]int{1, 0, 0, 0}},
					retui.NewStyle(),
					components.Button().
						ID("submit").
						Focused(isFocused(8)).
						Label("Submit").
						Style(retui.NewStyle().Background(retui.Red)).
						HoverStyle(retui.NewStyle().Background(retui.Green).Bold(true)).
						ActiveStyle(retui.NewStyle().Foreground(retui.White).Background(retui.Red).Bold(true)).
						OnClick(func(id string) {
							println("Button", id, "clicked!")
						}).
						Render(),
				),
			).
			Render(),
	)
}
