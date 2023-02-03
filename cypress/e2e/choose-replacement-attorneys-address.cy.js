import {AddressFormAssertions} from "../support/e2e";

describe('Choose replacement attorneys address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-replacement-attorneys-address?id=without-address&withIncompleteAttorneys=1');
    });

    it('address can be looked up', () => {
        AddressFormAssertions.assertCanAddAddressFromSelect()
        cy.url().should('contain', '/choose-replacement-attorneys-summary');
    });

    it('address can be entered manually if not found', () => {
        AddressFormAssertions.assertCanAddAddressManually('I can’t find their address in the list')
        cy.url().should('contain', '/choose-replacement-attorneys-summary');
    });

    it('errors when empty postcode', () => {
        AddressFormAssertions.assertErrorsWhenPostcodeEmpty()
    });

    it('errors when invalid postcode', () => {
        AddressFormAssertions.assertErrorsWhenInvalidPostcode()
    });

    it('errors when valid postcode and no addresses', () => {
        AddressFormAssertions.assertErrorsWhenValidPostcodeFormatButNoAddressesFound()
    });

    it('errors when unselected', () => {
        AddressFormAssertions.assertErrorsWhenUnselected()
    });

    it('errors when manual incorrect', () => {
        AddressFormAssertions.assertErrorsWhenManualIncorrect('I can’t find their address in the list')
    });
});
