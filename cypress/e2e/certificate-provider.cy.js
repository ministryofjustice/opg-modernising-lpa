describe('Certificate provider task', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list&withAttorney=1');
    });

    it('can be done later', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

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

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

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

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type('07535111111');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Online and by email').click();
        cy.get('#f-email').type('someone@example.com');
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Solicitor').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/check-your-lpa');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.visitLpa('/task-list')
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('can be a lay person', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type('07535111111');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Using paper forms').click();
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/certificate-provider-address');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').should('have.value', '2 RICHMOND PLACE');
        cy.get('#f-address-line-2').should('have.value', '');
        cy.get('#f-address-line-3').should('have.value', '');
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Friend').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-long-have-you-known-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How long have you known John Doe?');
        cy.contains('label', '2 years or more').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

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
        cy.contains('[for=f-email] + .govuk-error-message', 'Enter certificate provider\'s email address');

        cy.get('#f-email').type('not-an-email', { force: true });
        cy.contains('button', 'Continue').click()
        cy.contains('[for=f-email] + .govuk-error-message', 'Certificate provider\'s email address must be in the correct format, like name@example.com');
    });
    
    it('errors when empty postcode', () => {
        cy.visitLpa('/certificate-provider-address');

        cy.contains('button', 'Find address').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter postcode');
        });
        
        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter postcode');
    });

    it('errors when unselected', () => {
        cy.visitLpa('/certificate-provider-address');

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select address');
        });
        
        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select address');
    });

    it('errors when manual incorrect', () => {
        cy.visitLpa('/certificate-provider-address');

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', "Can not find address?").click();
        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter address line 1');
            cy.contains('Enter town or city');
        });
        
        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Enter address line 1');
        cy.contains('[for=f-address-town] + .govuk-error-message', 'Enter town or city');

        cy.get('#f-address-line-1').invoke('val', 'a'.repeat(51));
        cy.get('#f-address-line-2').invoke('val', 'b'.repeat(51));
        cy.get('#f-address-line-3').invoke('val', 'c'.repeat(51));
        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Address line 1 must be 50 characters or less');
        cy.contains('[for=f-address-line-2] + .govuk-error-message', 'Address line 2 must be 50 characters or less');
        cy.contains('[for=f-address-line-3] + .govuk-error-message', 'Address line 3 must be 50 characters or less');
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
});
