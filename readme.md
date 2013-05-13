#Clusterbrot

Clusterbrot is a mandlebrot set generator and renderer aimed for use on a distributed system. MScheduler.go will accept connections from MNode.go and send them a segment of the set to generate. The nodes calculate the given data and send it back. Once all the segments have been calculated, the Scheduler uses a color function to render the set as a png.
