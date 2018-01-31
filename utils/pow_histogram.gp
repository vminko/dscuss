# This script has three parameters:
#   infile  - input file with the data
#   outfile - output file (image)
#   bitnum  - number of zero bits in PoW
#   count   - number of samples in the input data

set xlabel 'T, sec'
set ylabel 'Frequency'

width=0.2
set boxwidth width*1
bin(x,width)=width*floor(x/width)+width/2.0

set style fill solid 0.4

set term png
set output outfile

plot infile using (bin($1,width)):(1.0) \
	smooth freq with boxes \
	title "\nTime to find ".bitnum."-bit PoW\n(".count." samples, binsize = 0.2 sec)" \
	linecolor rgb "#55AA55"
