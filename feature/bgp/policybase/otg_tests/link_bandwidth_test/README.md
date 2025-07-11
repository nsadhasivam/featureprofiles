# RT-7.5: BGP Policy - Match and Set Link Bandwidth Community

## Summary

Configure bgp policy to match, add and delete statically configured BGP link
bandwidth communities to routes based on a prefix match.

## Testbed type

* [2 port ATE to DUT](https://github.com/openconfig/featureprofiles/blob/main/topologies/atedut_2.testbed)

## Procedure

* Testbed configuration - Setup external BGP sessions and prefixes.
  * Generate config for 2 DUT and ATE ports where:
    * DUT port 1 to ATE port 1 EBGP session.
    * DUT port 2 to ATE port 2 IBGP session.
  * Configure dummy accept policies and attach it to both sessions on DUT.
     * Create a `/routing-policy/policy-definitions/policy-definition/policy-definition`
    named 'allow_all' with the following `statements`
       * statement[name='allow-all']/
         * actions/config/policy-result = ACCEPT_ROUTE
     * Use `/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/import-policy`
      to apply the policy on the DUT bgp neighbor to the ATE port 1.
  * Configure ATE port 1 with a BGP session to DUT port 1.
    * Advertise ipv4 and ipv6 prefixes to DUT port 1 using the following communities:
      * prefix-set-1 with 2 ipv4 and 2 ipv6 routes without communities.
      * prefix-set-2 with 2 ipv4 and 2 ipv6 routes with communities `[ "100:100" ]`.
      * [TODO value change] prefix-set-3 with 2 ipv4 and 2 ipv6 routes with extended communities `[ "link-bandwidth:23456:1000" ]`.
  * Configure Send community knob to IBGP neigbour to advertise the communities to IBGP peer 
    * [TODO] use `/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/config/send-community-type` leaf-list  with value `[STANDARD EXTENDED]`.
* RT-7.5.1 - Validate bgp sessions and traffic
  * For IPv4 and IPv6 prefixes:
    * Observe received prefixes at ATE port-2.
  * Generate traffic from ATE port-2 to ATE port-1.
  * Verify traffic is received on ATE port 1 for advertised prefixes.
    routes.

* RT-7.5.2 - Validate adding and removing link-bandwidth ext-community-sets using OC model release 3.x
  * Configure the following extended community sets on the DUT:
    (prefix: `routing-policy/defined-sets/bgp-defined-sets/ext-community-sets/ext-community-set`)
    * Create an ext-community-set named 'linkbw_1M' with members as follows:
      * ext-community-member = [ "link-bandwidth:23456:1M" ]
    * Create an ext-community-set named 'linkbw_2G' with members as follows:
      * ext-community-member = [ "link-bandwidth:23456:2G" ]
    * Create an community-set named 'regex_match_comm100' with members as follows:
      * community-member = [ "^100:.*$" ]
    * Create an ext-community-set named 'linkbw_any' with members as follows:
      * ext-community-member = [ "^link-bandwidth:.*:.*$" ]

  * Create a `/routing-policy/policy-definitions/policy-definition/policy-definition`
    named **'not_match_100_set_linkbw_1M'** with the following `statements`
    * statement[name='1-megabit-match']/
      * conditions/bgp-conditions/match-community-set/config/community-set = 'regex_match_comm100'
      * conditions/bgp-conditions/match-community-set/config/match-set-options = INVERT
      * actions/bgp-actions/set-ext-community/reference/config/ext-community-set-refs = 'linkbw_1M'
      * actions/config/policy-result = NEXT_STATEMENT
    * statement[name='accept_all_routes']/
      * actions/config/policy-result = ACCEPT_ROUTE

  * Create a `/routing-policy/policy-definitions/policy-definition/policy-definition`
    named **'match_100_set_linkbw_2G'** with the following `statements`
    * statement[name='2-gigabit-match']/
      * conditions/bgp-conditions/match-community-set/config/community-set = 'regex_match_comm100'
      * conditions/bgp-conditions/match-community-set/config/match-set-options = ANY
      * actions/bgp-actions/set-ext-community/reference/config/ext-community-set-refs = 'linkbw_2G'
      * actions/config/policy-result = NEXT_STATEMENT
    * statement[name='accept_all_routes']/
      * actions/config/policy-result = ACCEPT_ROUTE

  * Create a `/routing-policy/policy-definitions/policy-definition/policy-definition`
    named **'del_linkbw'** with the following `statements`
    * statement[name='del_linkbw']/
      * actions/bgp-actions/set-ext-community/config/options = 'REMOVE'
      * actions/bgp-actions/set-ext-community/config/method = 'REFERENCE'
      * actions/bgp-actions/set-ext-community/reference/config/ext-community-set-refs = 'linkbw_any'
      * actions/config/policy-result = NEXT_STATEMENT
    * statement[name='accept_all_routes']/
      * actions/config/policy-result = ACCEPT_ROUTE

  * For each policy-definition created, run a subtest (RT-7.5.3.x-<policy_name_here>-import) to
    * Use gnmi Set REPLACE option for:
      * `/routing-policy/policy-definitions` to configure the policy
      * Use `/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/import-policy`
        to apply the policy on the DUT bgp neighbor to the ATE port 1.
    * Verify expected communities are present in ATE.
    * Verify expected communities are present in DUT state.
      * Do not fail test if this path is not supported, only log results
        * `/network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes/route/state/ext-community-index`
        * `/network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-post/routes/route/state/ext-community-index`
      * Mark test as passing if Global Administartive valuee (ASN) of link-bakdwidth extended community **send by DUT** is either `23456` or ASN of DUT. 

    * Expected community values for each policy
      |              | set_linkbw_0                           | not_match_100_set_linkbw_1M |
      | ------------ | -------------------------------------- | --------------------------- |
      | prefix-set-1 | *DEPRECATED*                           | [ "link-bandwidth:23456:1000000" ] |
      | prefix-set-2 | *DEPRECATED*                           | [ "100:100" ]               |
      | prefix-set-3 | *DEPRECATED*                           | [ "link-bandwidth:23456:1000000" ] |

      |              | match_100_set_linkbw_2G                           | del_linkbw    | rm_any_zero_bw_set_LocPref_5 |
      | ------------ | ------------------------------------------------- | ------------- | ---------------------------- |
      | prefix-set-1 | [ none ]                                          | [none]        | *DEPRECATED*                 |
      | prefix-set-2 | [  "100:100", "link-bandwidth:23456:2000000000" ] | [ "100:100" ] | *DEPRECATED*                 |
      | prefix-set-3 | [ "link-bandwidth:23456:1000" ]                   | [ none ]      | *DEPRECATED*                 |

      * Regarding prefix-set-3 and policy "nomatch_100_set_linkbw_2G"
        * prefix-set-3 is advertised to the DUT with community "link-bandwidth:100:0" set.
        * The DUT evaluates a match for "regex_nomatch_as100".  This does not match because the regex pattern does not include the link-bandwidth community type.
        * Community linkbw_2G should be added.

[TODO] For each policy-definition created, run a subtest (RT-7.5.4.x-<policy_name_here>-export) to

  * Use gnmi Set REPLACE option to attach `allow_all` policy  using `/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/export-policy` to apply the policy on the DUT bgp neighbor to the ATE port 1.
  * Use gnmi Set REPLACE option for:
    * `/routing-policy/policy-definitions` to configure the policy
    * Use `/network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/export-policy`
        to apply the policy on the DUT bgp neighbor to the ATE port 2.
  * Verify expected communities are present in ATE port 2.
    * Mark test as passing if Global Administartive valuee (ASN) of link-bakdwidth extended community **send by DUT** is either `23456` or ASN of DUT. 

   * Expected community values for each policy

      |              | set_linkbw_0                           | not_match_100_set_linkbw_1M |
      | ------------ | -------------------------------------- | --------------------------- |
      | prefix-set-1 | *DEPRECATED*                           | [ "link-bandwidth:23456:1000000" ]  |
      | prefix-set-2 | *DEPRECATED*                           | [ "100:100" ]               |
      | prefix-set-3 | *DEPRECATED*                           | [ "link-bandwidth:23456:1000000" ]  |

      |              | match_100_set_linkbw_2G                           | del_linkbw    | rm_any_zero_bw_set_LocPref_5 |
      | ------------ | ------------------------------------------------- | ------------- | ---------------------------- |
      | prefix-set-1 | [ none ]                                          | [none]        | *DEPRECATED*                 |
      | prefix-set-2 | [  "100:100", "link-bandwidth:23456:2000000000" ] | [ "100:100" ] | *DEPRECATED*                 |
      | prefix-set-3 | [ "link-bandwidth:23456:1000" ]                   | [ none ]      | *DEPRECATED*                 |

  * Regarding prefix-set-3 and policy "nomatch_100_set_linkbw_2G"
    * prefix-set-3 is advertised to the DUT with community "link-bandwidth:100:0" set.
    * The DUT evaluates a match for "regex_nomatch_as100".  This does not match because the regex pattern does not include the link-bandwidth community type.
    * Community linkbw_2G should be added.

## OpenConfig Path and RPC Coverage

The below yaml defines the OC paths intended to be covered by this test.  OC paths used for test setup are not listed here.

```yaml
paths:
  ## Config Parameter Coverage
  ## Configuration to enable advertise communities to bgp peer
  /network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/config/send-community-type:
  ## Policy for community-set configuration
  /routing-policy/defined-sets/bgp-defined-sets/ext-community-sets/ext-community-set/config/ext-community-set-name:
  /routing-policy/defined-sets/bgp-defined-sets/ext-community-sets/ext-community-set/config/ext-community-member:
  ## Policy action configuration
  /routing-policy/policy-definitions/policy-definition/config/name:
  /routing-policy/policy-definitions/policy-definition/statements/statement/config/name:
  /routing-policy/policy-definitions/policy-definition/statements/statement/actions/config/policy-result:
  /routing-policy/policy-definitions/policy-definition/statements/statement/actions/bgp-actions/set-ext-community/config/options:
  /routing-policy/policy-definitions/policy-definition/statements/statement/actions/bgp-actions/set-community/config/method:
  /routing-policy/policy-definitions/policy-definition/statements/statement/actions/bgp-actions/set-ext-community/reference/config/ext-community-set-refs:
  /routing-policy/policy-definitions/policy-definition/statements/statement/actions/bgp-actions/config/set-local-pref:
  ## Policy for community-set match configuration
  /routing-policy/policy-definitions/policy-definition/statements/statement/conditions/bgp-conditions/match-ext-community-set/config/ext-community-set:
  /routing-policy/policy-definitions/policy-definition/statements/statement/conditions/bgp-conditions/match-ext-community-set/config/match-set-options:
  ## Policy attachment point configuration
  /network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/import-policy:
  /network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/afi-safis/afi-safi/apply-policy/config/export-policy:
  /network-instances/network-instance/protocols/protocol/bgp/peer-groups/peer-group/apply-policy/config/import-policy:
  /network-instances/network-instance/protocols/protocol/bgp/peer-groups/peer-group/apply-policy/config/export-policy:
  ## Telemetry Parameter Coverage
  /network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes/route/state/ext-community-index:
  /network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-post/routes/route/state/ext-community-index:
rpcs:
    gnmi:
        gNMI.Subscribe:
```

## Canonical OC
```json
{}
``` 

## Minimum DUT Required

vRX - Virtual Router Device
