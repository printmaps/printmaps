#!/usr/bin/perl -w
# Grayscale CSS:
# Replace all CSS colors from STDIN with their grayscale versions in the STDOUT.
#   Supports hex colors, rgb/rgba, and hsl/hsla.
# Usage: perl grayscale-css.pl screen.css > grayscale-screen.css
# by Weston Ruter <http://weston.ruter.net/>
# Copyright 2009, Shepherd Interactive <http://shepherdinteractive.com/>
# License: GPL 3.0 <http://www.gnu.org/licenses/gpl.html>
# 
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# 
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

use warnings;
use strict;

use open ':utf8';
our $VERSION = '1.0';


# Convert RGB(A) color into a grayscale version by adding "together 30% of the
# red value, 59% of the green value, and 11% of the blue value":
# http://en.wikipedia.org/wiki/Grayscale#Converting_color_to_grayscale
# Returns CSS string if not wantarray
sub grayscaleRGB {
	my($r,$g,$b,$a) = @_;
	$a = 1.0 if !defined $a;
	
	# Convert RGB to grayscale
	$r=$g=$b = 0.30*$r + 0.59*$g + 0.11*$b;
	
	return ($r,$g,$b,$a) if wantarray;
	
	# Return CSS color
	return $a == 1.0 ?
		sprintf("#%x%x%x", $r, $g, $b)
		:
		sprintf("rgba(%d, %d, %d, $a)", int($r + 0.5), int($g + 0.5), int($b + 0.5));
		# "rgba($r, $g, $b, $a)";
}


# Convert HSL(A) color into a grayscale version.
# Returns CSS string if not wantarray
sub grayscaleHSL {
	my($h,$s,$l,$a) = @_;
	$a = 1.0 if !defined $a;
	
	$s = 0; # "All to easy... Impressive. Most impressive." --Darth Vader
	
	return ($h,$s,$l,$a) if wantarray;
	
	# Return CSS color
	return $a == 1.0 ?
		"hsl($h, $s%, $l%)"
		:
		"hsla($h, $s%, $l%, $a)";
}


$_ = join '', <>;

#$_ = <<TEST;
##123
##AF4FCC;
#rgb(1,200,1)
#rgba(34,200,23,0.5)
#hsl(120, 100%, 50%)
#hsla(500, 50%, 70%, 0.5)
#TEST


# Normal 6-digit hex color
s{#([0-9A-F]{6})\b}
{
	scalar grayscaleRGB(
		hex substr($1, 0, 2),
		hex substr($1, 2, 2),
		hex substr($1, 4, 2)
	)
}eig;

# Shortcut 2-digit hex color
s{#([0-9A-F]{3})\b}
{
	scalar grayscaleRGB(
		hex substr($1, 0, 1)x2,
		hex substr($1, 1, 1)x2,
		hex substr($1, 2, 1)x2
	)
}eig;

# rgb()/rgba() color
s{
	rgba?\s*\(\s*
		(\d*\.?\d*)\s*,\s*
		(\d*\.?\d*)\s*,\s*
		(\d*\.?\d*)\s*
		(?:,\s*
			(\d*\.?\d*)\s*
		)?
	\)
}
{
	scalar grayscaleRGB(
		$1,
		$2,
		$3,
		$4
	)
}eigx;

# hsl()/hsla() color
s{
	hsla?\s*\(\s*
		(\d*\.?\d*)\s*,\s*
		(\d*\.?\d*)%\s*,\s*
		(\d*\.?\d*)%\s*
		(?:,\s*
			(\d*\.?\d*)\s*
		)?
	\)
}
{
	scalar grayscaleHSL(
		$1,
		$2,
		$3,
		$4
	)
}eigx;

print;
