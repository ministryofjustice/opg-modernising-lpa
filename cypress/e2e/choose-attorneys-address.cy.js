import {AddressFormAssertions} from "../support/e2e";

describe('Choose attorneys address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-attorneys-address?id=without-address&withIncompleteAttorneys=1');
    });

    it('address can be looked up', () => {
        AddressFormAssertions.assertCanAddAddressFromSelect()
        cy.url().should('contain', '/choose-attorneys-summary');
    });

    it('address can be entered manually if not found', () => {
        AddressFormAssertions.assertCanAddAddressManually('I can’t find their address in the list')
        cy.url().should('contain', '/choose-attorneys-summary');
    });

    it('address can be entered manually on invalid postcode', () => {
        AddressFormAssertions.assertCanAddAddressManually('Enter address manually', true)
        cy.url().should('contain', '/choose-attorneys-summary');
    });

    it('errors when empty postcode', () => {
        AddressFormAssertions.assertErrorsWhenPostcodeEmpty()
    });

    it('errors when unselected', () => {
        AddressFormAssertions.assertErrorsWhenUnselected()
    });

    it('errors when manual incorrect', () => {
        AddressFormAssertions.assertErrorsWhenManualIncorrect('I can’t find their address in the list')
    });
});
