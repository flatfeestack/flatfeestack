<?php
$ldap_host = "ldap://127.0.0.1:8389";


$ds = ldap_connect($ldap_host)
         or exit(">>Could not connect to LDAP server<<");
ldap_set_option($ds, LDAP_OPT_PROTOCOL_VERSION, 3);
//ldap_start_tls($ds) ;
ldap_bind($ds,"tom","1234") ;
// sleep(10);
ldap_close($ds);

echo "ok\n";
