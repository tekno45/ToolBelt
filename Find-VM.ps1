function Find-VM{

param
([string]$ComputerName)

$servers = "SCVMMCSS2", "OAKITVMM", "OAKENGSCVMM2"

foreach($server in $servers){   

$scvmmserv = Get-SCVMMServer $server
$VM = Get-SCVirtualMachine $ComputerName


"i can string wherever i want"

if($VM.Name){Return $VM, $scvmmserv} 


Else{"No VM"}

}


}