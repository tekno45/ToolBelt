function Find-VM{

param
([string]$ComputerName)

#List of SCVMM servers to query
$servers = "SCVMMCSS2", "OAKITVMM", "OAKENGSCVMM2"

#Loop through servers and search for machine on each
foreach($server in $servers){   

#Capture server and VM info
$scvmmserv = Get-SCVMMServer $server
$VM = Get-SCVirtualMachine $ComputerName

#Once found return info
if($VM.Name){Return $VM, $scvmmserv} 

}
#NO VM FOUND 
Return "No VM"
}