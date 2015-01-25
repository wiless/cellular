#!/usr/bin/env python
import socket
import struct
import pylab
from pylab import *
import time
import array
import matplotlib.pyplot as plt
import numpy as np

def get_udp_packets():
	UDP_IP = '127.0.0.1'
	UDP_PORT = 8080
	BUFFER_SIZE = 2100  # Normally 1024, but we want fast response
	sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
	sock.bind((UDP_IP, UDP_PORT))
	#sock.settimeout(1.0)
	header=struct.Struct('< 10s d d q')
	pklen=0
	i=1
	packet_data=[]
	total_len=0
	while True:
		try:
			packet, addr = sock.recvfrom(BUFFER_SIZE) # buffer size is 1024 bytes
		#print packet
			pklen=len(packet)
			print 'Receiving packet number = ',i,' with length =', pklen
			i=i+1
			
			if pklen>0:
				sock.settimeout(1.0)
				rec_data=list(header.unpack(packet[:34]))
				print rec_data[1],rec_data[-1]
				total_len+=rec_data[-1]
				Data_format=struct.Struct('<%dd' % rec_data[-1])
				data=array.array('d',packet[34:])
				values=list(Data_format.unpack_from(data))			
				packet_data.extend(values)
		#print 'packet==',packet
		except socket.timeout:
			print 'total number of values rec=',total_len
			#print 'CODE will break'
			break
	#print 'CODE is OUT'
	#print len(Header_list)
	#exit()	
  #print "Packet size:", len(packet)
	
  #print " Header:",rec_data
	
	
	return packet_data
  #print values 



def main():
	i=0
	fig, ax = plt.subplots()
	line, = ax.plot(np.random.randn(256))
	plt.ion()
	plt.ylim([0,1])
	plt.ion()
	plt.show(block=False)
	ax.set_ylabel('Amplitude')
	ax.set_xlabel('Sample number')
	while True:
		udp_data=get_udp_packets()
		print '-'*100
		print 'Reception number=',i
		print 'Total number of values =',len(udp_data)
		window=256*5
		No_plots = len(udp_data)/window
		print No_plots
		for j in xrange(No_plots):
			axleng = window
			xmin=j*axleng
			xmax=(j+1)*axleng
			xax=range(xmin,xmax)
			line.set_ydata(udp_data[j*window:(j+1)*window])
			line.set_xdata(xax)
			plt.xlim(xmin,xmax)
			ax.draw_artist(ax.patch)
			ax.draw_artist(line)
			fig.canvas.update()
			fig.canvas.flush_events()
		i+=1

main()