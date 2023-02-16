import {AddressFormAssertions} from "../support/e2e";

describe('People to notify address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?withIncompletePeopleToNotify=1&redirect=/choose-people-to-notify-address?id=JoannaSmith');
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
