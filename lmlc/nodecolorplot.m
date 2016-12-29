%load nodeinfo.dat % NodeID,Freq,X,Y,SINR
%nodeinfo =nodeinfo(find(mod(nodeinfo(:,7),2)==0),:);

nodeinfo=nodecolortable;
frequency= unique(nodeinfo(:,2))';
for f=frequency
nodeinfoTable=nodeinfo(find(nodeinfo(:,2)==f),:);

figure
col=5
sinr=nodeinfoTable(:,col);
cmap=colormap;
LEVELS=length(cmap);
minsinr=-32;
maxsinr=50;
sinrrange=(maxsinr-minsinr);
cedges=[0:LEVELS-1]*sinrrange/LEVELS+(minsinr);

clevel=quantiz(sinr,cedges);

N=length(nodeinfoTable);
 	deltasize=80/14;
	S=80*ones(N,1);
sinrrange
LEVELS
delta = (sinrrange/LEVELS)
C=floor(sinr/delta);
C=cedges(clevel);
scatter3(nodeinfoTable(:,3),nodeinfoTable(:,4),nodeinfoTable(:,col),S,C,'filled')

colorbar
view(2)
title(f)
end
hold on
plot(bslocations(:,2),bslocations(:,3),'*k','MarkerSize',10)
hold on;
plot(antennalocations(:,1),antennalocations(:,2),'Or','MarkerSize',10) 

% stable=stable(1:500,:);
bestbsid=stable(:,7);

drawPolyGon(complex(bslocations(:,2),bslocations(:,3)),ISD/2);
drawPolyGon(complex(antennalocations(:,1),antennalocations(:,2)),ISD/2,'b');
 
