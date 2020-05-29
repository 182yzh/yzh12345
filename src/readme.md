// 529 有问题
&{Pass 6214e9 application_1506638472019_16565 [{2017-10-09 03:07:23 None [{m146 [gpu7]}]} {2017-10-09 03:11:14 2017-10-09 03:11:57 [{m146 [gpu7]}]} {2017-10-09 03:14:24 None [{m146 [gpu7]}]} {None 2017-10-09 03:15:12 [{m153 [gpu1]}]}] 2017-10-09 03:04:16 2c46d5 }


// 1303 之间有一个也有问题
803 &{Killed 6214e9 application_1506638472019_15647 [] 2017-10-08 16:10:00 2c46d5 }
好多killed


忽略有多个task的job
解析相应的job ，计算出其需要的gpu以及submit time，run_time,进行模拟，到submit提交，运
行结束之后进行处理。（定义json对象，解析到相应的结构）
（按submit time 排队）

描述到达 调度 结束
记录相关信息：在任务结束或者任务到达（任务需求超过资源时，不进行处理）
3. 当前排队的任务数，及每一个任务的
   gpu需求
   任务的等待时间
   任务的开始时间
1. 新任务到达，任务结束，是否有任务被调度（任务的信息（GPU需求），被调度到某个节点的>信息）
2. 节点的GPU使用情况

==--== 代表可能还会有问题，debug应该着重观看


//question
attempts 有很多次的是以最后一次为准？
有的任务可能运行时间会比较长emmm（可能会有数天）

res这边暂时是全部自己维护？
gang schedule的相关问题


graphmanager 以及 flow scheduler 主要涉及点的加入，以及少量边的管理（
目前只包括unagg node to sink)

在添加每一个任务以及job 后，通知cs（costmodel），进行处理（包含选择相应的res，是否连接unaggenode）
添加res之后，由cost model 给出res to node 的边


problem:如果一个task完成之后，如何处理？（交由cs？）
目前考虑是：gm处理的部分封装（removetasknode）
然后交由cs进行处理（目前cs拟采用的思路：更改root task的supply以及sinknode的supply）
cs的调用由flow scheduler 进行实现


mind: 在每次由cs进行处理taskcomplete之后，检查是否jobcomplete  or 每次更新时检查jobcomplete
mind: task complete之后，需要对res descriptor 中的current running task -删掉task （交由flow scheduler实现）


question: 如果考虑图本身较小，是不是可以直接通过每次重建图？（怎么样都要去建图，初始的图需要，另一方面，如果不考虑抢占的问题，实际上并不需要已经在运行的任务参与，他们起到的作用只是占用了一部分资源，完全可以考虑----只有需要调度的job以及task）


question : rd.currentRunningTask由谁来加


mind: task state 注意更新

question： 如task被调度，是否应该更改连接到unschdule agg node的边（或者修改容量，或者修改代价）
        node被调度后的更新问题。
        

思路：在cost model 对所有的task 添加一个root task，并将所有的task的流（gpu resquest）并入root task
（如果在task 进入时进行更新，可能会使一个task的流被加多次？addtask，updatetask（多次），）





问题： 在处理数量级到达一次调度200（需求8，4，2，1，各50个）明显感觉计算需要的时间变长（350个2 gpu Res，50个8gpu）

还没实现：
修改部分cs2.exe 删去没有流量的边的输出。（本身优化较小）
限定每个tasknode 的prefer resNode 数量（存在一个问题：如何保证task的prefer res不会集中与一部分，而是尽可能的均匀分布）

在不支持抢占的情况下，当集群中的running task逐渐增多时，也会给cs2.exe 增加负担
暂时考虑：为了维持平衡，只把其他的边删除，但是running task的边不再去限定容量一定为1（通过只有这一条边通过这个流量，实现绑定）


// ju he yibufen ---
next: 

1. ping jun deng dai shi jian de fen bu
2. every machine  ulization/  zong ti de ziyuanliyonglu
3. ziyuanliyonglv cdf / total cdf(time)
4. waiting queue with time change.  res li yong lv  sui shijian bian hua (python shuju chuli)
5. zai bu tong de gpu need xia de waitting time / waiting queue zhong de gpu xuqiu shu mu

// xuyao   submit time change  then  new ---  
// xuyao   change res kind 8.16.3
// xuyao 



1. 
10-03 00:00 kaishi jilu shijian /
10-03zhiqian xinagdangyu yure  zheer dui ziyuan liyong lv de yingxiang .

2. chuqu 10-03 00:00 zhiqiande/12-15 18:42 zhihoude  renwu ,kanyiyan ziyuanliyonglv 

3. 


https://www.bilibili.com/blackboard/topic/activity-OObFWXZI8.html