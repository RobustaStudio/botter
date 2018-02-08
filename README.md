Botter
========
> just a server that compiles html file to messenger component, so from now you can write messenger bot using html components

Why
======
> at first it was a for fun project, to just find a simple way to write a chat bot for collecting data without repeating any logic each time, so we started using a json file then moved to yaml and finally xml (some html components).

Status
=======
> This is the beta release and it is being re-written from scratch to clean its code-base and optimize it as much as possible so we can add more features in the future.

How it works
=============
> Simply, it takes an xml/html file, then builds a so-called `AST` tree after that it converts each component to a messenger component to be send to the messenger thread, then for each thread/session it creates a session storage using redis to store the collected data, after reaching the end of each form, it will submit the submitted data to a backend from the form `action` attribute just like the basic html forms.

> In another words, it uses the concept of a browser engine, but instead of compiling to the operating system components it compiles to messenger components.

Features
=========
- Standalone binary with redis as session storage
- Compiles common `HTML` components to its Messenger equivalent
- It allows you to write custom templates to respond to the user based on defined words!
- It can fetch the response from a backend, we called it `composer`.

Chat bot example
================
```html
<html version="1.0" id="test">
	<head>
		<meta name="title" content="Testing Bot" />
		<meta name="description" content="Some description about this bot here" />
		<meta name="app-secret" content="*******" />
		<meta name="verify-secret" content="false" />
		<meta name="verify-token" content="123456" />
		<meta name="page-token" content="****************" />
		<meta name="error" content="Sorry, but I really cannot find what you need right now, please try again or use the navigator" />
		<meta name="composer" content="http://localhost:8000/composer.php?q=%s" />
	</head>
	<body>
		<nav id="main" title="Welcom to the main nav">
			<a href="nav://intro" reset="true">Intro</a>
			<a href="form://about">About Me</a>
		</nav>
		<nav id="intro" title="Welcom to the intro nav">
			<a href="form://info">Send Your Info</a>
			<a href="https://google.com" embed="false" ratio="full">Visit Google</a>
			<a href="nav://main" reset="true">Back to Home</a>
		</nav>
		<form id="about" title="about - lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum lorem ipsum"></form>
		<form id="info" method="post" action="https://requestb.in/17uvpbb1" title="follow the following steps ..." submit="thank you ^_^!">
			<input id="type" type="options" title="What is your type :)?">
				<option value="company">Company</option>
				<option value="personal">Personal</option>
			</input>
			<input if="type==company" type="text" id="company_name" title="Please write the company name"/>
			<input if="type==personal" type="text" id="name" title="Please write the your name"/>
			<input type="file" id="brief" title="Please upload a brief about yourself (pdf, doc, docx)"/>
		</form>
		<template match=".*">
			<!-- audio, video, image, file -->
			<!-- it will select a random reply from those: -->
			<reply label="some text here" />
			<reply label="wow" type="image" src="https://laravel.com/favicon.png"/>
		</template>
	</body>
</html>
```

Installation
============
- using docker ? - `docker run --network host alash3al/botter -http :8080 -bot "/path/to/chatbot.html" -redis "redis://localhost:6379/10"`
- portable binary ? - goto [releases](https://github.com/RobustaStudio/botter/releases) page and select yours.
- building from source ? - `go get -u github.com/RobustaStudio/botter`

Credits
========
[Robustastudio](https://robustastudio.com)

License
==========
[MIT License](LICENSE)
