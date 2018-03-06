function Find-VM{

param
(
[parameter(Mandatory=$true, Position=0)][string]$ComputerName,
[pscredential]$credentials,
$servers = (Import-Csv -Path C:\users\jgreene\Documents\ToolBelt\SCVMM_Servers.csv),
[switch]$Full
)

#Loop through servers and search for machine on each
foreach($server in $servers){   


#Capture server and VM info
    $scvmmserv = Get-SCVMMServer $server.Server -Credential $credentials
    $VM = Get-SCVirtualMachine -All -VMMServer $scvmmserv | `
    Where-Object{$_.ComputerName -eq "$ComputerName.osisoft.int" -or $_.ComputerName -eq "$computerName.dev.osisoft.int"}

#Once found return info

if($VM){
#Experimental flags, options for quick status and for full VM object
    $status = "Currently "+$VM.Status,"in " + $scvmmserv.Name + "environment `n", "CPU at "+ $VM.CPUUtilization + " percent", "with " + $VM.MemoryAvailablePercentage + " percent memory Available"
    Return $status

    if($Full){
        Return $VM
    }
} 

}
#NO VM FOUND 
Return "No VM"
}