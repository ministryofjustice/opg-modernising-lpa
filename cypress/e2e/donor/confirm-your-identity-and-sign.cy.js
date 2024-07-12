describe('Confirm your identity and sign', () => {
  describe('when certificate provider is acting online', () => {
    beforeEach(() => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
    });

    it('can be completed ', () => {
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Sam Smith (donor)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/onelogin-identity-details');
      cy.checkA11yApp();

      cy.contains('Sam');
      cy.contains('Smith');
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/read-your-lpa');
      cy.checkA11yApp();

      cy.contains('h2', "LPA decisions");
      cy.contains('h2', "People named on the LPA");
      cy.contains('h3', "Donor");
      cy.contains('h3', "Attorneys");
      cy.contains('h3', "Replacement attorney");
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/your-lpa-language');
      cy.contains('label', 'Continue and register my LPA in English').click();
      cy.contains('button', 'Save and continue').click();

      cy.url().should('contain', '/your-legal-rights-and-responsibilities');
      cy.checkA11yApp();
      cy.contains('a', 'Continue to signing page').click();

      cy.url().should('contain', '/sign-your-lpa');
      cy.checkA11yApp();

      cy.contains('h1', "Sign your LPA");
      cy.contains('label', 'I want to sign this LPA as a deed').click();
      cy.contains('label', 'I want to apply to register this LPA').click();
      cy.contains('button', 'Submit my signature').click();

      cy.url().should('contain', '/witnessing-your-signature');
      cy.checkA11yApp();

      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/witnessing-as-certificate-provider');
      cy.checkA11yApp();

      cy.contains('h1', "Confirm you witnessed the donor sign");
      cy.get('#f-witness-code').type('1234');
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/you-have-submitted-your-lpa');
      cy.checkA11yApp();

      cy.contains('h1', "You’ve submitted your LPA");
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/dashboard');
    });

    it('errors when not signed', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=confirmYourIdentity');

      cy.visitLpa('/sign-your-lpa');

      cy.contains('button', 'Submit my signature').click();

      cy.get('.govuk-error-summary').within(() => {
        cy.contains('Select both boxes to sign and apply to register your LPA');
      });

      cy.contains('.govuk-error-message', 'Select both boxes to sign and apply to register your LPA');
    });

    it('errors when not witnessed', () => {
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('a', 'Continue').click();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Sam Smith (donor)').click();
      cy.contains('button', 'Continue').click();

      cy.contains('button', 'Continue').click();
      cy.contains('a', 'Continue').click();
      cy.contains('label', 'Continue and register my LPA in English').click();
      cy.contains('button', 'Save and continue').click();
      cy.contains('a', 'Continue to signing page').click();
      cy.contains('label', 'I want to sign this LPA as a deed').click();
      cy.contains('label', 'I want to apply to register this LPA').click();
      cy.contains('button', 'Submit my signature').click();

      cy.contains('button', 'Continue').click();
      cy.contains('button', 'Continue').click();

      cy.get('.govuk-error-summary').within(() => {
        cy.contains('Enter the code we sent to the certificate provider');
      });

      cy.contains('.govuk-error-message', 'Enter the code we sent to the certificate provider');

      cy.get('#f-witness-code').type('123');
      cy.contains('button', 'Continue').click();

      cy.contains('.govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');

      cy.get('#f-witness-code').type('45');
      cy.contains('button', 'Continue').click();

      cy.contains('.govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');
    });
  })

  describe('when certificate provider is acting on paper', () => {
    it('can be completed and paper forms are requested', () => {
      cy.visit('/fixtures?redirect=/read-your-lpa&progress=confirmYourIdentity&certificateProvider=paper');

      cy.url().should('contain', '/read-your-lpa');
      cy.checkA11yApp();

      cy.contains('h2', "LPA decisions");
      cy.contains('h2', "People named on the LPA");
      cy.contains('h3', "Donor");
      cy.contains('h3', "Attorneys");
      cy.contains('h3', "Replacement attorney");
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/your-lpa-language');
      cy.contains('label', 'Continue and register my LPA in English').click();
      cy.contains('button', 'Save and continue').click();

      cy.url().should('contain', '/your-legal-rights-and-responsibilities');
      cy.checkA11yApp();
      cy.contains('a', 'Continue to signing page').click();

      cy.url().should('contain', '/sign-your-lpa');
      cy.checkA11yApp();

      cy.contains('h1', "Sign your LPA");
      cy.contains('label', 'I want to sign this LPA as a deed').click();
      cy.contains('label', 'I want to apply to register this LPA').click();
      cy.contains('button', 'Submit my signature').click();

      cy.url().should('contain', '/witnessing-your-signature');
      cy.checkA11yApp();

      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/witnessing-as-certificate-provider');
      cy.checkA11yApp();

      cy.contains('h1', "Confirm you witnessed the donor sign");
      cy.get('#f-witness-code').type('1234');
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/you-have-submitted-your-lpa');
      cy.checkA11yApp();

      cy.contains('h1', "You’ve submitted your LPA");
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/dashboard');

      cy.contains('.govuk-body-s', 'Reference number:')
        .invoke('text')
        .then((text) => {
          const uid = text.split(':')[1].trim();
          cy.visit(`http://localhost:9001/?detail-type=paper-form-requested&detail=${uid}`);

          cy.contains(`"uid":"${uid}"`)
          cy.contains(`"actorType":"certificateProvider"`)
        });
    });
  })

  describe('when insufficient evidence to prove identity', () => {
    it('can start vouching journey', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Unable to prove identity (X)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/unable-to-confirm-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/what-is-vouching');
      cy.checkA11yApp();
    })
  })

  describe('when any other return code', () => {
    it('shows problem', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Failed identity check (T)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/register-with-court-of-protection');
      cy.checkA11yApp();
      cy.contains('register the LPA through the Court of Protection');

      cy.contains('a', 'Return to task list').click();
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'There is a problem')
        .find('a')
        .click();

      cy.url().should('contain', '/register-with-court-of-protection');
    })

    it('can withdraw', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Failed identity check (T)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/register-with-court-of-protection');
      cy.checkA11yApp();
      cy.contains('register the LPA through the Court of Protection');

      cy.contains('label', 'I no longer want to make this LPA').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/withdraw-this-lpa');
    })
  })

  describe('when identity details do not match LPA', () => {
    it('can update LPA details', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Charlie Cooper (certificate provider)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/onelogin-identity-details');
      cy.checkA11yApp();

      cy.contains('dd', 'Sam').parent().contains('span', 'Does not match');
      cy.contains('dd', 'Smith').parent().contains('span', 'Does not match');
      cy.contains('dd', '2 January 2000').parent().contains('span', 'Does not match');

      cy.contains('label', 'Yes').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/onelogin-identity-details');
      cy.checkA11yApp();

      cy.contains('Your LPA details have been updated to match your confirmed identity')
      cy.get('main').should('not.contain', 'Sam');
      cy.get('main').should('not.contain', 'Smith');
      cy.get('main').should('not.contain', '2 January 2000');
    })

    it('can withdraw LPA', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Charlie Cooper (certificate provider)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/onelogin-identity-details');
      cy.checkA11yApp();

      cy.contains('dd', 'Sam').parent().contains('span', 'Does not match');
      cy.contains('dd', 'Smith').parent().contains('span', 'Does not match');
      cy.contains('dd', '2 January 2000').parent().contains('span', 'Does not match');

      cy.contains('label', 'No').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/withdraw-this-lpa');
      cy.checkA11yApp();
    })

    it('errors when option not selected', () => {
      cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa');
      cy.contains('li', "Confirm your identity and sign")
        .should('contain', 'Not started')
        .find('a')
        .click();

      cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
      cy.checkA11yApp();

      cy.contains('h1', 'How to confirm your identity and sign the LPA');
      cy.contains('a', 'Continue').click();

      cy.url().should('contain', '/prove-your-identity');
      cy.checkA11yApp();
      cy.contains('a', 'Continue').click();

      cy.contains('label', 'Charlie Cooper (certificate provider)').click();
      cy.contains('button', 'Continue').click();

      cy.url().should('contain', '/onelogin-identity-details');
      cy.checkA11yApp();

      cy.contains('button', 'Continue').click();

      cy.get('.govuk-error-summary').within(() => {
        cy.contains('Select yes if you would like to update your details');
      });

      cy.contains('.govuk-error-message', 'Select yes if you would like to update your details');
    });
  });
});
