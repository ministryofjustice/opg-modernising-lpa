import { AddressFormAssertions } from "../../support/e2e";

describe('Donor address', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/your-address');
    });

    it('address can be looked up', () => {
        AddressFormAssertions.assertCanAddAddressFromSelect()
        cy.url().should('contain', '/receiving-updates-about-your-lpa');
    });

    it('address can be entered manually if not found', () => {
        AddressFormAssertions.assertCanAddAddressManually('I can’t find my address in the list')
        cy.url().should('contain', '/receiving-updates-about-your-lpa');
    });

    it('address can be entered manually on invalid postcode', () => {
        AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)
        cy.url().should('contain', '/receiving-updates-about-your-lpa');
    });

    it('errors when empty postcode', () => {
        AddressFormAssertions.assertErrorsWhenYourPostcodeEmpty()
    });

    it('errors when invalid postcode', () => {
        AddressFormAssertions.assertErrorsWhenInvalidPostcode()
    });

    it('errors when valid postcode and no addresses', () => {
        AddressFormAssertions.assertErrorsWhenYourValidPostcodeFormatButNoAddressesFound()
    });

    it('errors when unselected', () => {
        AddressFormAssertions.assertErrorsWhenYourAddressUnselected()
    });

    it('errors when manual incorrect', () => {
        AddressFormAssertions.assertErrorsWhenYourManualIncorrect('I can’t find my address in the list')
    });
});
