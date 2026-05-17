# Engage tools — execution N/A matrix

Auto-generated. Regenerate:

```bash
python3 scripts/engage/generate-tools-na-matrix.py
make test-engage-na-matrix
```

**Catalog tools:** 158 | **Live enabled (tools.live.yaml):** 138 | **Catalog ∩ live:** 48

| Tool | Binary | Category | Status | Reason |
|------|--------|----------|--------|--------|
| `advanced_payload_generation` | `advanced` | web | bridge_api | workflow placeholder binary `advanced` |
| `ai_generate_attack_suite` | `ai` | web | bridge_api | workflow placeholder binary `ai` |
| `ai_generate_payload` | `ai` | web | bridge_api | workflow placeholder binary `ai` |
| `ai_reconnaissance_workflow` | `ai` | web | bridge_api | workflow placeholder binary `ai` |
| `ai_test_payload` | `ai` | web | bridge_api | workflow placeholder binary `ai` |
| `ai_vulnerability_assessment` | `ai` | web | bridge_api | workflow placeholder binary `ai` |
| `amass_scan` | `amass` | osint | live | enabled in tools.live.yaml |
| `analyze_target_intelligence` | `analyze` | intelligence | bridge_api | workflow placeholder binary `analyze` |
| `anew_data_processing` | `anew` | web | runner_N/A | binary not in engage-runner image |
| `angr_symbolic_execution` | `angr` | web | live | enabled in tools.live.yaml |
| `api_fuzzer` | `api` | web | bridge_api | in-process MCP bridge handler |
| `api_schema_analyzer` | `api` | intelligence | bridge_api | in-process MCP bridge handler |
| `arjun_parameter_discovery` | `arjun` | web | live | enabled in tools.live.yaml |
| `arjun_scan` | `arjun` | web | live | enabled in tools.live.yaml |
| `arp_scan_discovery` | `arp` | web | runner_N/A | binary not in engage-runner image |
| `autorecon_comprehensive` | `autorecon` | web | bridge_api | workflow placeholder binary `autorecon` |
| `autorecon_scan` | `autorecon` | web | bridge_api | workflow placeholder binary `autorecon` |
| `binwalk_analyze` | `binwalk` | binary | live | enabled in tools.live.yaml |
| `browser_agent_inspect` | `browser` | web | bridge_api | workflow placeholder binary `browser` |
| `bugbounty_authentication_bypass_testing` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_business_logic_testing` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_comprehensive_assessment` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_file_upload_testing` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_osint_gathering` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_reconnaissance_workflow` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `bugbounty_vulnerability_hunting` | `bugbounty` | intelligence | bridge_api | workflow placeholder binary `bugbounty` |
| `burpsuite_alternative_scan` | `burpsuite` | web | live | enabled in tools.live.yaml |
| `burpsuite_scan` | `burpsuite` | web | live | enabled in tools.live.yaml |
| `checkov_iac_scan` | `checkov` | web | bridge_api | workflow placeholder binary `checkov` |
| `checksec_analyze` | `checksec` | intelligence | bridge_api | workflow placeholder binary `checksec` |
| `clair_vulnerability_scan` | `clair` | web | bridge_api | workflow placeholder binary `clair` |
| `clear_cache` | `clear` | web | bridge_api | workflow placeholder binary `clear` |
| `cloudmapper_analysis` | `cloudmapper` | cloud | bridge_api | workflow placeholder binary `cloudmapper` |
| `comprehensive_api_audit` | `comprehensive` | web | bridge_api | workflow placeholder binary `comprehensive` |
| `correlate_threat_intelligence` | `correlate` | intelligence | runner_N/A | binary not in engage-runner image |
| `create_attack_chain_ai` | `create` | web | bridge_api | workflow placeholder binary `create` |
| `create_file` | `create` | web | bridge_api | workflow placeholder binary `create` |
| `create_scan_summary` | `create` | web | bridge_api | workflow placeholder binary `create` |
| `create_vulnerability_report` | `create` | web | bridge_api | workflow placeholder binary `create` |
| `ctf_auto_solve_challenge` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_binary_analyzer` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_create_challenge_workflow` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_cryptography_solver` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_forensics_analyzer` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_suggest_tools` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `ctf_team_strategy` | `api` | ctf | bridge_api | in-process MCP bridge handler |
| `dalfox_xss_scan` | `dalfox` | web | live | enabled in tools.live.yaml |
| `delete_file` | `delete` | web | runner_N/A | binary not in engage-runner image |
| `detect_technologies_ai` | `detect` | web | runner_N/A | binary not in engage-runner image |
| `dirb_scan` | `dirb` | web | live | enabled in tools.live.yaml |
| `dirsearch_scan` | `dirsearch` | web | live | enabled in tools.live.yaml |
| `discover_attack_chains` | `discover` | web | runner_N/A | binary not in engage-runner image |
| `display_system_metrics` | `display` | web | runner_N/A | binary not in engage-runner image |
| `dnsenum_scan` | `dnsenum` | web | live | enabled in tools.live.yaml |
| `docker_bench_security_scan` | `docker` | web | runner_N/A | binary not in engage-runner image |
| `dotdotpwn_scan` | `dotdotpwn` | web | runner_N/A | binary not in engage-runner image |
| `enum4linux_ng_advanced` | `enum4linux` | web | live | enabled in tools.live.yaml |
| `enum4linux_scan` | `enum4linux` | web | live | enabled in tools.live.yaml |
| `error_handling_statistics` | `error` | web | runner_N/A | binary not in engage-runner image |
| `execute_command` | `execute` | web | bridge_api | workflow placeholder binary `execute` |
| `execute_python_script` | `engage-python-exec` | web | live | enabled in tools.live.yaml |
| `exiftool_extract` | `exiftool` | web | runner_N/A | binary not in engage-runner image |
| `falco_runtime_monitoring` | `falco` | web | runner_N/A | binary not in engage-runner image |
| `feroxbuster_scan` | `feroxbuster` | web | live | enabled in tools.live.yaml |
| `ffuf_scan` | `ffuf` | web | live | enabled in tools.live.yaml |
| `fierce_scan` | `fierce` | web | live | enabled in tools.live.yaml |
| `foremost_carving` | `foremost` | web | runner_N/A | binary not in engage-runner image |
| `format_tool_output_visual` | `format` | web | runner_N/A | binary not in engage-runner image |
| `gau_discovery` | `gau` | web | live | enabled in tools.live.yaml |
| `gdb_analyze` | `gdb` | binary | live | enabled in tools.live.yaml |
| `gdb_peda_debug` | `gdb` | binary | live | enabled in tools.live.yaml |
| `generate_exploit_from_cve` | `generate` | web | bridge_api | workflow placeholder binary `generate` |
| `generate_payload` | `generate` | web | bridge_api | workflow placeholder binary `generate` |
| `get_cache_stats` | `get` | web | bridge_api | workflow placeholder binary `get` |
| `get_live_dashboard` | `get` | web | bridge_api | workflow placeholder binary `get` |
| `get_process_dashboard` | `get` | web | bridge_api | workflow placeholder binary `get` |
| `get_process_status` | `get` | web | bridge_api | workflow placeholder binary `get` |
| `get_telemetry` | `get` | web | bridge_api | workflow placeholder binary `get` |
| `ghidra_analysis` | `ghidra` | binary | live | enabled in tools.live.yaml |
| `gobuster_scan` | `gobuster` | web | live | enabled in tools.live.yaml |
| `graphql_scanner` | `graphql` | web | runner_N/A | binary not in engage-runner image |
| `hakrawler_crawl` | `hakrawler` | web | runner_N/A | binary not in engage-runner image |
| `hashcat_crack` | `hashcat` | auth | live | enabled in tools.live.yaml |
| `hashpump_attack` | `hashpump` | web | runner_N/A | binary not in engage-runner image |
| `http_framework_test` | `http` | web | bridge_api | workflow placeholder binary `http` |
| `http_intruder` | `http` | web | bridge_api | workflow placeholder binary `http` |
| `http_repeater` | `http` | web | bridge_api | workflow placeholder binary `http` |
| `http_set_rules` | `http` | web | bridge_api | workflow placeholder binary `http` |
| `http_set_scope` | `http` | web | bridge_api | workflow placeholder binary `http` |
| `httpx_probe` | `httpx` | web | live | enabled in tools.live.yaml |
| `hydra_attack` | `hydra` | auth | live | enabled in tools.live.yaml |
| `install_python_package` | `engage-python-install` | web | live | enabled in tools.live.yaml |
| `intelligent_smart_scan` | `intelligent` | web | runner_N/A | binary not in engage-runner image |
| `jaeles_vulnerability_scan` | `jaeles` | web | live | enabled in tools.live.yaml |
| `john_crack` | `john` | auth | live | enabled in tools.live.yaml |
| `jwt_analyzer` | `jwt` | intelligence | runner_N/A | binary not in engage-runner image |
| `katana_crawl` | `katana` | web | live | enabled in tools.live.yaml |
| `kube_bench_cis` | `kube` | cloud | bridge_api | workflow placeholder binary `kube` |
| `kube_hunter_scan` | `kube` | cloud | bridge_api | workflow placeholder binary `kube` |
| `libc_database_lookup` | `libc` | web | runner_N/A | binary not in engage-runner image |
| `list_active_processes` | `list` | web | bridge_api | workflow placeholder binary `list` |
| `list_files` | `list` | web | bridge_api | workflow placeholder binary `list` |
| `masscan_high_speed` | `masscan` | network | live | enabled in tools.live.yaml |
| `metasploit_run` | `metasploit` | web | live | enabled in tools.live.yaml |
| `modify_file` | `modify` | web | runner_N/A | binary not in engage-runner image |
| `monitor_cve_feeds` | `monitor` | web | runner_N/A | binary not in engage-runner image |
| `msfvenom_generate` | `msfvenom` | web | runner_N/A | binary not in engage-runner image |
| `nbtscan_netbios` | `nbtscan` | web | live | enabled in tools.live.yaml |
| `netexec_scan` | `netexec` | web | runner_N/A | binary not in engage-runner image |
| `nikto_scan` | `nikto` | web | live | enabled in tools.live.yaml |
| `nmap_advanced_scan` | `nmap` | network | live | enabled in tools.live.yaml |
| `nmap_scan` | `nmap` | network | live | enabled in tools.live.yaml |
| `nuclei_scan` | `nuclei` | web | live | enabled in tools.live.yaml |
| `objdump_analyze` | `objdump` | intelligence | runner_N/A | binary not in engage-runner image |
| `one_gadget_search` | `one` | web | runner_N/A | binary not in engage-runner image |
| `optimize_tool_parameters_ai` | `optimize` | web | runner_N/A | binary not in engage-runner image |
| `pacu_exploitation` | `pacu` | web | runner_N/A | binary not in engage-runner image |
| `paramspider_discovery` | `paramspider` | web | live | enabled in tools.live.yaml |
| `paramspider_mining` | `paramspider` | web | live | enabled in tools.live.yaml |
| `pause_process` | `pause` | web | runner_N/A | binary not in engage-runner image |
| `prowler_scan` | `prowler` | cloud | runner_N/A | binary not in engage-runner image |
| `pwninit_setup` | `pwninit` | web | runner_N/A | binary not in engage-runner image |
| `pwntools_exploit` | `pwntools` | web | runner_N/A | binary not in engage-runner image |
| `qsreplace_parameter_replacement` | `qsreplace` | web | runner_N/A | binary not in engage-runner image |
| `radare2_analyze` | `radare2` | binary | live | enabled in tools.live.yaml |
| `research_zero_day_opportunities` | `research` | web | runner_N/A | binary not in engage-runner image |
| `responder_credential_harvest` | `responder` | web | runner_N/A | binary not in engage-runner image |
| `resume_process` | `resume` | web | runner_N/A | binary not in engage-runner image |
| `ropgadget_search` | `ropgadget` | web | runner_N/A | binary not in engage-runner image |
| `ropper_gadget_search` | `ropper` | web | runner_N/A | binary not in engage-runner image |
| `rpcclient_enumeration` | `rpcclient` | web | runner_N/A | binary not in engage-runner image |
| `rustscan_fast_scan` | `rustscan` | network | live | enabled in tools.live.yaml |
| `scout_suite_assessment` | `scout` | cloud | runner_N/A | binary not in engage-runner image |
| `select_optimal_tools_ai` | `select` | web | runner_N/A | binary not in engage-runner image |
| `server_health` | `server` | web | runner_N/A | binary not in engage-runner image |
| `smbmap_scan` | `smbmap` | web | runner_N/A | binary not in engage-runner image |
| `sqlmap_scan` | `sqlmap` | web | live | enabled in tools.live.yaml |
| `steghide_analysis` | `steghide` | web | runner_N/A | binary not in engage-runner image |
| `strings_extract` | `strings` | web | runner_N/A | binary not in engage-runner image |
| `subfinder_scan` | `subfinder` | osint | live | enabled in tools.live.yaml |
| `target_timeline_intelligence` | `api` | intelligence | bridge_api | in-process MCP bridge handler |
| `terminate_process` | `terminate` | web | runner_N/A | binary not in engage-runner image |
| `terrascan_iac_scan` | `terrascan` | web | runner_N/A | binary not in engage-runner image |
| `test_error_recovery` | `test` | web | runner_N/A | binary not in engage-runner image |
| `threat_hunting_assistant` | `threat` | web | runner_N/A | binary not in engage-runner image |
| `trivy_scan` | `trivy` | cloud | live | enabled in tools.live.yaml |
| `uro_url_filtering` | `uro` | web | runner_N/A | binary not in engage-runner image |
| `volatility3_analyze` | `volatility3` | intelligence | runner_N/A | binary not in engage-runner image |
| `volatility_analyze` | `volatility` | intelligence | live | enabled in tools.live.yaml |
| `vulnerability_intelligence_dashboard` | `vulnerability` | intelligence | runner_N/A | binary not in engage-runner image |
| `wafw00f_scan` | `wafw00f` | web | live | enabled in tools.live.yaml |
| `waybackurls_discovery` | `waybackurls` | web | live | enabled in tools.live.yaml |
| `wfuzz_scan` | `wfuzz` | web | runner_N/A | binary not in engage-runner image |
| `wpscan_analyze` | `wpscan` | web | live | enabled in tools.live.yaml |
| `x8_parameter_discovery` | `x8` | web | live | enabled in tools.live.yaml |
| `xsser_scan` | `xsser` | web | runner_N/A | binary not in engage-runner image |
| `xxd_hexdump` | `xxd` | web | runner_N/A | binary not in engage-runner image |
| `zap_scan` | `zap` | web | runner_N/A | binary not in engage-runner image |
