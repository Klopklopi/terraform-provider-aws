// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.LicenseManagerServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"You have reached the maximum allowed number of license configurations created in one day",
	)
}

func TestAccLicenseManagerLicenseConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var licenseConfiguration licensemanager.GetLicenseConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_license_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLicenseConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "license-manager", regexache.MustCompile(`license-configuration:lic-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "license_count", acctest.CtZero),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "false"),
					resource.TestCheckResourceAttr(resourceName, "license_counting_type", "Instance"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.#", acctest.CtZero),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtZero),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLicenseManagerLicenseConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var licenseConfiguration licensemanager.GetLicenseConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_license_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLicenseConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflicensemanager.ResourceLicenseConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLicenseManagerLicenseConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var licenseConfiguration licensemanager.GetLicenseConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_license_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLicenseConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseConfigurationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLicenseConfigurationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtTwo),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLicenseConfigurationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLicenseManagerLicenseConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var licenseConfiguration licensemanager.GetLicenseConfigurationOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_license_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LicenseManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLicenseConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseConfigurationConfig_allAttributes(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "license-manager", regexache.MustCompile(`license-configuration:lic-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test1"),
					resource.TestCheckResourceAttr(resourceName, "license_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "true"),
					resource.TestCheckResourceAttr(resourceName, "license_counting_type", "Socket"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "license_rules.0", "#minimumSockets=3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtZero),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLicenseConfigurationConfig_allAttributesUpdated(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLicenseConfigurationExists(ctx, resourceName, &licenseConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "license-manager", regexache.MustCompile(`license-configuration:lic-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test2"),
					resource.TestCheckResourceAttr(resourceName, "license_count", "99"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "false"),
					resource.TestCheckResourceAttr(resourceName, "license_counting_type", "Socket"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "license_rules.0", "#minimumSockets=3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.CtZero),
				),
			},
		},
	})
}

func testAccCheckLicenseConfigurationExists(ctx context.Context, n string, v *licensemanager.GetLicenseConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No License Manager License Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn(ctx)

		output, err := tflicensemanager.FindLicenseConfigurationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLicenseConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_licensemanager_license_configuration" {
				continue
			}

			_, err := tflicensemanager.FindLicenseConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("License Manager License Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLicenseConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Instance"
}
`, rName)
}

func testAccLicenseConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Instance"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLicenseConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Instance"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccLicenseConfigurationConfig_allAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                     = %[1]q
  description              = "test1"
  license_count            = 10
  license_count_hard_limit = true
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]
}
`, rName)
}

func testAccLicenseConfigurationConfig_allAttributesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  description           = "test2"
  license_count         = 99
  license_counting_type = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]
}
`, rName)
}
