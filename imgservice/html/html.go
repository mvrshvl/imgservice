package html

const (
	Home = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Обработка изображений</title>
 </head>
 <body>
  <form action="/resize-percent" target="_blank">
   <button>Resize (n%)</button>
  </form>
  <form action="/resize" target="_blank">
   <button>Resize (width/height)</button>
  </form>
  <form action="/convert" target="_blank">
   <button>Convert</button>
  </form>
  <form action="/grayscale" target="_blank">
   <button>Gray scale</button>
  </form>
  <form action="/watermark" target="_blank">
   <button>Watermark</button>
  </form>
 </body>
</html>`

	Resize = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Resize</title>
 </head>
 <body>
  <form method="post" action="/resize-percent/load" enctype="multipart/form-data">
	  <p><b>Размер в процентах:</b><br>
	   <input name="size" type="text" size="40">
	  </p>
	 <div>
	   <label for="file">Choose file to upload</label>
	   <input type="file" id="file" name="file" multiple>
	 </div>
	 <div>
	   <button>Submit</button>
	 </div>
  </form>
 </body>
</html>`
)
