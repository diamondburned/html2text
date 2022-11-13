package html2text

import (
	"fmt"
)

func Example() {
	inputHTML := `
<html>
	<head>
		<title>My Mega Service</title>
		<link rel=\"stylesheet\" href=\"main.css\">
		<style type=\"text/css\">body { color: #fff; }</style>
	</head>

	<body>
		<div class="logo">
			<a href="http://jaytaylor.com/"><img src="/logo-image.jpg" alt="Mega Service"/></a>
		</div>

		<h1>Welcome to your new account on my service!</h1>

		<p>
			Here is some more information:

			<ul>
				<li>Link 1: <a href="https://example.com">Example.com</a></li>
				<li>Link 2: <a href="https://example2.com">Example2.com</a></li>
				<li>Something else</li>
			</ul>
		</p>

		<table>
			<thead>
				<tr><th>Header 1</th><th>Header 2</th></tr>
			</thead>
			<tfoot>
				<tr><td>Footer 1</td><td>Footer 2</td></tr>
			</tfoot>
			<tbody>
				<tr><td>Row 1 Col 1</td><td>Row 1 Col 2</td></tr>
				<tr><td>Row 2 Col 1</td><td>Row 2 Col 2</td></tr>
			</tbody>
		</table>
	</body>
</html>`

	text, err := FromString(inputHTML, Options{PrettyTables: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(text)

	// Output:
	// Mega Service ( http://jaytaylor.com/ )
	//
	// ******************************************
	// Welcome to your new account on my service!
	// ******************************************
	//
	// Here is some more information:
	//
	// * Link 1: Example.com ( https://example.com )
	// * Link 2: Example2.com ( https://example2.com )
	// * Something else
	//
	// +-------------+-------------+
	// |  HEADER 1   |  HEADER 2   |
	// +-------------+-------------+
	// | Row 1 Col 1 | Row 1 Col 2 |
	// | Row 2 Col 1 | Row 2 Col 2 |
	// +-------------+-------------+
	// |  FOOTER 1   |  FOOTER 2   |
	// +-------------+-------------+
}
