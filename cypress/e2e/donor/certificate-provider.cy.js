import {
    AddressFormAssertions, TestEmail, TestMobile
} from "../../support/e2e";

describe('Certificate provider task', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list&withDonorDetails=1&withAttorney=1');
    });

    it('can be done later', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('button', 'I will do this later').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started');
    });

    it('can be left unfinished', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.visitLpa('/task-list');

        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'In progress');
    });

    it('can be a professional', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.checkA11yApp();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type(TestMobile);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Online and by email').click();
        cy.get('#f-email').type(TestEmail);
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Solicitor').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.visitLpa('/task-list')
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('can be a lay person', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.checkA11yApp();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type(TestMobile);
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Using paper forms').click();
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/certificate-provider-address');

        AddressFormAssertions.assertCanAddAddressFromSelect()

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Friend').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-long-have-you-known-certificate-provider');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How long have you known John Doe?');
        cy.contains('label', '2 years or more').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.visitLpa('/task-list')
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('errors when details empty', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter date of birth');
            cy.contains('Enter mobile number');
        });

        cy.contains('[for=f-first-names] + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
        cy.contains('[for=f-mobile] + p + .govuk-error-message', 'Enter mobile number');
    });

    it('errors when invalid mobile number', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.get('#f-mobile').type('not-a-number');
        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-mobile] + p + .govuk-error-message', 'Mobile number must be a UK mobile number, like 07700 900 982 or +44 7700 900 982');
    });

    it('errors when invalid dates of birth', () => {
        cy.visitLpa('/certificate-provider-details');

        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').clear().type('1990');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
    });

    it('errors when how they prefer to carry out their role unselected', () => {
        cy.visitLpa('/how-would-certificate-provider-prefer-to-carry-out-their-role');

        cy.contains('button', 'Continue').click()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how your certificate provider would prefer to carry out their role');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how your certificate provider would prefer to carry out their role');
    });

    it('errors when how they prefer to carry out their role email invalid', () => {
        cy.visitLpa('/how-would-certificate-provider-prefer-to-carry-out-their-role');

        cy.contains('label', 'Online and by email').click();
        cy.contains('button', 'Continue').click()
        cy.contains('[for=f-email] + .govuk-error-message', 'Enter certificate provider’s email address');

        cy.get('#f-email').type('not-an-email', { force: true });
        cy.contains('button', 'Continue').click()
        cy.contains('[for=f-email] + .govuk-error-message', 'Certificate provider’s email address must be in the correct format, like name@example.com');
    });

    it('errors when empty postcode', () => {
        cy.visitLpa('/certificate-provider-address');
        AddressFormAssertions.assertErrorsWhenPostcodeEmpty()
    });

    it('errors when unselected', () => {
        cy.visitLpa('/certificate-provider-address');
        AddressFormAssertions.assertErrorsWhenUnselected()
    });

    it('errors when manual incorrect', () => {
        cy.visitLpa('/certificate-provider-address');
        AddressFormAssertions.assertErrorsWhenManualIncorrect('I can’t find their address in the list')
    });

    it('errors when invalid postcode', () => {
        cy.visitLpa('/certificate-provider-address');
        AddressFormAssertions.assertErrorsWhenInvalidPostcode()
    });

    it('errors when how you know not selected', () => {
        cy.visitLpa('/how-do-you-know-your-certificate-provider');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how you know your certificate provider');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how you know your certificate provider');
    });

    it('errors relationship not explained', () => {
        cy.visitLpa('/how-do-you-know-your-certificate-provider');

        cy.contains('label', 'Other').click();
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter description');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Enter description');
    });

    it('errors how long you have known them not selected', () => {
        cy.visitLpa('/how-long-have-you-known-certificate-provider');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how long you have known your certificate provider');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how long you have known your certificate provider');
    });

    it('errors when known for less than 2 years', () => {
        cy.visitLpa('/how-long-have-you-known-certificate-provider');

        cy.contains('label', 'Less than 2 years').click();
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('You must have known your non-professional certificate provider for 2 years or more');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'You must have known your non-professional certificate provider for 2 years or more');
    });

    it('warns when name shared with other actor', () => {
        cy.visitLpa('/certificate-provider-details');
        cy.contains('button', 'Continue').click();

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Smith');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-mobile').type(TestMobile);
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.contains('There is also an attorney called John Smith.');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
    });
});
