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



		<h1>よおこそ！ Welcome to your new account on my service!</h1>



		<p>
			Here is some more information:

			<ul>
				<li>Link 1: <a href="https://example.com">Example.com</a></li>
				<li>Link 2: <a href="https://example2.com">Example2.com</a></li>
				<li>Something else</li>
			</ul>

			Here's a really long paragraph:
			<br />
			Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Nascetur ridiculus mus mauris vitae. Est ante in nibh mauris cursus mattis. Amet volutpat consequat mauris nunc. Eget egestas purus viverra accumsan in nisl nisi scelerisque. Id aliquet risus feugiat in ante metus. Sit amet commodo nulla facilisi nullam vehicula ipsum. Feugiat in ante metus dictum at tempor commodo. Tincidunt tortor aliquam nulla facilisi cras fermentum odio eu feugiat. Pharetra sit amet aliquam id.
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
	// Mega Service (http://jaytaylor.com/)
	//
	// *****************************************************
	// よおこそ！ Welcome to your new account on my service!
	// *****************************************************
	//
	// Here is some more information:
	//
	// - Link 1: Example.com (https://example.com)
	// - Link 2: Example2.com (https://example2.com)
	// - Something else
	//
	// Here's a really long paragraph:
	//
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
	// incididunt ut labore et dolore magna aliqua. Nascetur ridiculus mus mauris
	// vitae. Est ante in nibh mauris cursus mattis. Amet volutpat consequat mauris
	// nunc. Eget egestas purus viverra accumsan in nisl nisi scelerisque. Id aliquet
	// risus feugiat in ante metus. Sit amet commodo nulla facilisi nullam vehicula
	// ipsum. Feugiat in ante metus dictum at tempor commodo. Tincidunt tortor
	// aliquam nulla facilisi cras fermentum odio eu feugiat. Pharetra sit amet
	// aliquam id.
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
