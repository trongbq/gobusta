# Gobusta

Gobusta is a handly tool for building mininal blog site.

### Usage
```bash
gobusta <command> [arguments]

Usage:
	gobusta <command> [arguments]

The commands are:

	build 	build and render entire source
	serve	start a local server which serve generated files

And few options
	build --clean to clean the build directory before build content
	serve --port [int] to specify the running port number
```

### Blog structure
Please check the example directory to know a basic structure for your site.
Generally `gobusta` will assume there are three directories in your site repository
- `content` contains content for your site, which is a markdown type format.
- `templates` contains the template to render posts, there are two type of templates which are `post.html` for post content and `index.html` for the main page.
- `static` is a directory for assets file like Javascripts or CSS

### Front Matter
With current version, only two data on front matter which are
- `title` is the title of the post
- `date` is the published date
- `tags` contains list of tag

The format of a post will be something like this

```
{
  "title": "Gobusta, the blog engine",
  "date": "2020-05-11",
  "tags": ["sample", "post", "gobusta", "markdown"]
}
+++
Gobusta is a handly tool for building mininal blog site.
...
```

### License
Please see the [LICENSE](https://github.com/trongbq/gobusta/blob/master/LICENSE) file.
