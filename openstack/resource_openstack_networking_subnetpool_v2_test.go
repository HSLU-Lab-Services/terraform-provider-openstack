package openstack

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/subnetpools"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNetworkingV2SubnetPool_Basic(t *testing.T) {
	var subnetPool subnetpools.SubnetPool

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckNonAdminOnly(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNetworkingV2SubnetPoolDestroy(t.Context()),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2SubnetPoolBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2SubnetPoolExists(t.Context(), "openstack_networking_subnetpool_v2.subnetpool_1", &subnetPool),
					testAccCheckNetworkingV2SubnetPoolPrefixesConsistency("openstack_networking_subnetpool_v2.subnetpool_1", &subnetPool),
				),
			},
			{
				Config: testAccNetworkingV2SubnetPoolPrefixLengths,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "name", "subnetpool_1"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "description", "terraform subnetpool acceptance test"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "default_quota", "4"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "default_prefixlen", "25"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "min_prefixlen", "24"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "max_prefixlen", "30"),
				),
			},
			{
				Config: testAccNetworkingV2SubnetPoolUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "name", "subnetpool_1"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "description", "terraform subnetpool acceptance test updated"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "default_quota", "8"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "default_prefixlen", "26"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "min_prefixlen", "25"),
					resource.TestCheckResourceAttr(
						"openstack_networking_subnetpool_v2.subnetpool_1", "max_prefixlen", "28"),
				),
			},
		},
	})
}

func testAccCheckNetworkingV2SubnetPoolExists(ctx context.Context, n string, subnetPool *subnetpools.SubnetPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		networkingClient, err := config.NetworkingV2Client(ctx, osRegionName)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack networking client: %w", err)
		}

		found, err := subnetpools.Get(ctx, networkingClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return errors.New("Subnetpool not found")
		}

		*subnetPool = *found

		return nil
	}
}

func testAccCheckNetworkingV2SubnetPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		networkingClient, err := config.NetworkingV2Client(ctx, osRegionName)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack networking client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "openstack_networking_subnetpool_v2" {
				continue
			}

			_, err := subnetpools.Get(ctx, networkingClient, rs.Primary.ID).Extract()
			if err == nil {
				return errors.New("Subnetpool still exists")
			}
		}

		return nil
	}
}

func testAccCheckNetworkingV2SubnetPoolPrefixesConsistency(n string, subnetpool *subnetpools.SubnetPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ID is set")
		}

		for i, prefix := range subnetpool.Prefixes {
			if prefix != rs.Primary.Attributes[fmt.Sprintf("prefixes.%d", i)] {
				return errors.New("prefixes list elements or order is not consistent")
			}
		}

		return nil
	}
}

const testAccNetworkingV2SubnetPoolBasic = `
resource "openstack_networking_subnetpool_v2" "subnetpool_1" {
	name = "subnetpool_1"
	description = "terraform subnetpool acceptance test"

	prefixes = ["10.10.0.0/16", "10.11.11.0/24"]

	default_quota = 4

	default_prefixlen = 25
	min_prefixlen = 24
	max_prefixlen = 30
}
`

const testAccNetworkingV2SubnetPoolPrefixLengths = `
resource "openstack_networking_subnetpool_v2" "subnetpool_1" {
	name = "subnetpool_1"
	description = "terraform subnetpool acceptance test"

	prefixes = ["10.10.0.0/16", "10.11.11.0/24"]

	default_quota = 4

	default_prefixlen = 25
	min_prefixlen = 24
	max_prefixlen = 30
}
`

const testAccNetworkingV2SubnetPoolUpdate = `
resource "openstack_networking_subnetpool_v2" "subnetpool_1" {
	name = "subnetpool_1"
	description = "terraform subnetpool acceptance test updated"

	prefixes = ["10.10.0.0/16", "10.11.11.0/24"]

	default_quota = 8

	default_prefixlen = 26
	min_prefixlen = 25
	max_prefixlen = 28
}
`
