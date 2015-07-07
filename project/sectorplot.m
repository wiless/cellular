   close all
load table400.dat
load uelocations.dat 
load bslocations.dat
% 
 load antennalocations.dat
 antennalocations=antennalocations(:,2:end);


%x=fileread('antennaArray.json');
%antennas=JSON.parse(x);
%antennalocations=[];
 
%for k=1:length(antennas)
%antennalocations(k,:)=[antennas{k}.SettingAAS.Centre.X antennas{k}.SettingAAS.Centre.Y];
%end

stable400=sortrows(table400,1);
rows=length(stable400);
stable400=[stable400 uelocations(1:rows,2) uelocations(1:rows,3) angle(uelocations(1:rows,2)+i*uelocations(1:rows,3))*180/pi];

figure
% sp=plot(stable400(:,10),stable400(:,11),'m.')
% for k=1:(length(bslocations)/3)
% plot(uelocations([1:400]+400*(k-1),2),uelocations([1:400]+400*(k-1),3),'.');
% hold all
% end
syssinr=stable400(find(stable400(:,7)<57),8);
% % FILTER positive SINR users only
     stable400=stable400(find(stable400(:,8)>-3),:);
      %stable400=stable400(find(stable400(:,8)<=10),:)
figure 
cdfplot(syssinr)
figure(1)
% [Nrows Ncols]=size(stable400);
% NUEsPerCell=100;
% cell=3;
% uerows=[1:NUEsPerCell]+NUEsPerCell*(cell-1);
% stable400(Nrows,Ncols+3)=0;
% for indx=1:Nrows
%     findx=find(uelocations(:,1)==stable400(indx,1));
%     extracols=[uelocations(findx,2:3) radtodeg(angle(uelocations(findx,2)+i*uelocations(findx,3)))];
%     
%     stable400(indx,Ncols+1:Ncols+3)=extracols;
% end
% 
hold on;

plot(bslocations(:,2),bslocations(:,3),'*k','MarkerSize',10)
hold on;
plot(antennalocations(:,1),antennalocations(:,2),'Or','MarkerSize',10) 

% stable400=stable400(1:500,:);
bestbsid=stable400(:,7);

drawPolyGon(complex(bslocations(:,2),bslocations(:,3)),500);
drawPolyGon(complex(antennalocations(:,1),antennalocations(:,2)),500,'b');
 

nSectors=3; 
nCells=length(bslocations)/nSectors;
nCells=19
hold on;
k=0:nCells-1;
selectedUEs0=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec0ues=stable400(selectedUEs0,10:11);
k=k+nCells;
selectedUEs1=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec1ues=stable400(selectedUEs1,10:11);
k=k+nCells;
selectedUEs2=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec2ues=stable400(selectedUEs2,10:11);
h=plot(sec0ues(:,1),sec0ues(:,2),'r*');hold on
 plot(sec1ues(:,1),sec1ues(:,2),'k*');hold on
 plot(sec2ues(:,1),sec2ues(:,2),'b*');hold on
 legend 'sec0','sec1','sec2'

k=k+nCells;
selectedUEs0=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec0ues= stable400(selectedUEs0,10:11);
k=k+nCells;
selectedUEs1=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec1ues=stable400(selectedUEs1,10:11);
k=k+nCells;
selectedUEs2=find((bestbsid>=k(1)).*(bestbsid<=k(end)));
sec2ues=stable400(selectedUEs2,10:11);

h=plot(sec0ues(:,1),sec0ues(:,2),'m*');hold on
 plot(sec1ues(:,1),sec1ues(:,2),'g*');hold on
 plot(sec2ues(:,1),sec2ues(:,2),'c*');hold on
 legend 'sec00','sec01','sec02'


grid on