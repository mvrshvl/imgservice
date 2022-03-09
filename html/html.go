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

	ResizePercent = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Resize</title>
 </head>
 <body>
  <form method="post" action="/resize-percent/load" enctype="multipart/form-data">
	  <p><b>Размер в процентах:</b><br>
	   <input name="size" type="text" size="10">
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

	Download = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
 <head>
  <meta http-equiv="content-type" content="text/html; charset=utf-8">
  <title>Download file</title>
 </head>
 <body>
  <p><a href="localhost:6060/download/%s">Скачать изображения</a></p>
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
  <form method="post" action="/resize/load" enctype="multipart/form-data">
	  <p><b>Размер в пикселях:</b><br>
	   <input name="width" type="text" size="10">
	   <input name="height" type="text" size="10">
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

	GrayScale = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Resize</title>
 </head>
 <body>
  <form method="post" action="/grayscale/load" enctype="multipart/form-data">
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

	Convert = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Resize</title>
 </head>
 <body>
  <form method="post" action="/convert/load" enctype="multipart/form-data">
	 <select id="format" name="format">
	  <option value="png">PNG</option>
	  <option value="jpeg">JPEG</option>
	  <option value="pdf">PDF</option>
	  <option value="tiff">TIFF</option>
	  <option value="bmp">BMP</option>
	  <option value="gif">GIF</option>
	 </select>
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

	Watermark = `
<!DOCTYPE html>
<html>
 <head>
  <meta charset="utf-8">
  <title>Resize</title>
 </head>
 <body>
  <form method="post" action="/watermark/load" enctype="multipart/form-data">
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
