#!/usr/bin/python
# -*- coding: utf-8 -*-

# Nik4: Export image from mapnik
# Run it with -h to see the list of options
# Written by Ilya Zverev, licensed WTFPL

# Modifications for Printmaps project (Klaus Tockloth, 2017/05/12):
# - unused options removed
# - programme simplified
# - 16K resize limit removed
# - option "--info" added
# - size as float

import mapnik
import sys, os, re, argparse, math, tempfile

try:
	import cairo
	HAS_CAIRO = True
except ImportError:
	HAS_CAIRO = False

VERSION = '1.6 (Printmaps)'

p3857 = mapnik.Projection('+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +no_defs +over')
p4326 = mapnik.Projection('+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs')
transform = mapnik.ProjTransform(p4326, p3857)

def select_layers(m, enable, disable):
	"""Enable and disable layers in corresponding lists"""
	for l in m.layers:
		if l.name in enable:
			l.active = True
		if l.name in disable:
			l.active = False

if __name__ == "__main__":
	parser = argparse.ArgumentParser(description='Nik4 {}: mapnik image renderer'.format(VERSION))
	parser.add_argument('--version', action='version', version='Nik4 {}'.format(VERSION))
	parser.add_argument('--ppi', required=True, type=float, help='Pixels per inch (alternative to scale)')
	parser.add_argument('--scale', required=True, type=float, help='Scale as in 1:1000 (set ppi is recommended)')
	parser.add_argument('--size', required=True, nargs=2, metavar=('W', 'H'), type=float, help='Target dimensions in mm')
	parser.add_argument('--center', required=True, nargs=2, metavar=('X', 'Y'), type=float, help='Center of an image')
	parser.add_argument('--add-layers', help='Map layers to include, comma-separated')
	parser.add_argument('--hide-layers', help='Map layers to hide, comma-separated')
	parser.add_argument('--debug', action='store_true', default=False, help='Display calculated values')
	parser.add_argument('--info', action='store_true', default=False, help='Quit after displaying calculated values')
	parser.add_argument('style', help='Style file for mapnik')
	parser.add_argument('output', help='Resulting image file')
	options = parser.parse_args()

	dim_mm = None
	scale = None
	size = None
	bbox = None

	# format should not be empty
	if '.' in options.output:
		fmt = options.output.split('.')[-1].lower()
	else:
		fmt = 'png256'
	
	need_cairo = fmt in ['svg', 'pdf']
	
	# image size in millimeters
	dim_mm = options.size

	# ppi and scale factor are the same thing
	ppmm = options.ppi / 25.4
	scale_factor = options.ppi / 90.7

	# svg / pdf can be scaled only in cairo mode
	if scale_factor != 1 and need_cairo and not HAS_CAIRO:
		sys.stderr.write('Warning: install pycairo for using --ppi')
		scale_factor = 1
		ppmm = 90.7 / 25.4

	# convert physical size to pixels
	size = [int(round(dim_mm[0] * ppmm)), int(round(dim_mm[1] * ppmm))]

	if size and size[0] + size[1] <= 0:
		raise Exception('Both dimensions are less or equal to zero')

	# scale 1:NNN, has to divide by cos(lat)
	scale = options.scale * 0.00028 / scale_factor
	scale = scale / math.cos(math.radians(options.center[1]))

	# calculate bbox through center and target size
	center = transform.forward(mapnik.Coord(*options.center))
	w = size[0] * scale / 2
	h = size[1] * scale / 2
	bbox = mapnik.Box2d(center.x-w, center.y-h, center.x+w, center.y+h)

	# reading style xml into memory for preprocessing
	with open(options.style, 'r') as style_file:
		style_xml = style_file.read()
	style_path = os.path.dirname(options.style)

	# for layer processing we need to create the Map object
	m = mapnik.Map(size[0], size[1])
	mapnik.load_map_from_string(m, style_xml, False, style_path)
	m.srs = p3857.params()

	# add / remove some layers
	if options.add_layers or options.hide_layers:
		select_layers(m, options.add_layers.split(',') if options.add_layers else [], options.hide_layers.split(',') if options.hide_layers else [])

	if options.debug:
		print('scale={}'.format(scale))
		print('scale_factor={}'.format(scale_factor))
		print('size={},{}'.format(size[0], size[1]))
		print('bbox={}'.format(bbox))
		print('bbox_wgs84={}'.format(transform.backward(bbox) if bbox else None))
		print('layers=' + ','.join([l.name for l in m.layers if l.active]))

	if options.info:
	    quit()

	# export image
	m.aspect_fix_mode = mapnik.aspect_fix_mode.GROW_BBOX;
	m.zoom_to_box(bbox)

	outfile = options.output
	if options.output == '-':
		outfile = tempfile.TemporaryFile(mode='w+b')

	if need_cairo:
		if HAS_CAIRO:
			surface = cairo.SVGSurface(outfile, size[0], size[1]) if fmt == 'svg' else cairo.PDFSurface(outfile, size[0], size[1])
			mapnik.render(m, surface, scale_factor, 0, 0)
			surface.finish()
		else:
			mapnik.render_to_file(m, outfile, fmt)
	else:
		im = mapnik.Image(size[0], size[1])
		mapnik.render(m, im, scale_factor)
		im.save(outfile, fmt)

