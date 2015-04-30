load nodeinfo.dat % NodeID,X,Y,SINR
frequency= unique(nodeinfo(:,2))';
for f=frequency
nodeinfoTable=nodeinfo(find(nodeinfo(:,2)==f),:);

figure
sinr=nodeinfoTable(:,5);
cmap=colormap;
LEVELS=length(cmap);
minsinr=min(sinr);
maxsinr=max(sinr);
sinrrange=(maxsinr-minsinr);
% cedges=[0:LEVELS-1]*sinrrange/LEVELS+(minsinr);

% clevel=quantiz(sinr,cedges);

N=length(nodeinfoTable);
 	deltasize=80/14;
	S=80*ones(N,1);
    

scatter3(nodeinfoTable(:,3),nodeinfoTable(:,4),nodeinfoTable(:,5),S,floor(sinr/(sinrrange/LEVELS)),'filled')
colorbar
view(2)
end