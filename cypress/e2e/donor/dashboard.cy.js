describe('Dashboard', () => {
  context('with incomplete LPA', () => {
    beforeEach(() => {
      cy.visit('/fixtures/dashboard?asDonor=1&redirect=/dashboard');
    });

    it('shows my lasting power of attorney', () => {
      cy.checkA11yApp();

      cy.contains('Property and affairs');
      cy.contains('Sam Smith');
      cy.contains('strong', 'Drafting');
      cy.contains('a', 'Go to task list').click();

      cy.url().should('contain', '/task-list');
    });

    it('can create another LPA', () => {
      cy.contains('button', 'Start now').click();

      cy.url().should('contain', '/make-a-new-lpa');
      cy.checkA11yApp();
    });
  })

  context('with completed LPA', () => {
    it('completed LPAs have a track progress button', () => {
      cy.visit('/fixtures?redirect=&progress=signTheLpa');

      cy.get('button').should('not.contain', 'Continue');

      cy.contains('Property and affairs');
      cy.contains('Sam Smith');
      cy.contains('strong', 'In progress');
      cy.contains('a', 'Track LPA progress').click();

      cy.url().should('contain', '/progress');
    });
  });

  context('with perfect LPA', () => {
    it('shows the correct options', () => {
      cy.visit('/fixtures?redirect=&progress=submitted');

      cy.contains('Property and affairs');
      cy.contains('Sam Smith');
      cy.contains('strong', 'In progress');
      cy.contains('a', 'View LPA');
      cy.contains('a', 'Track LPA progress');
      cy.contains('a', 'Withdraw LPA');
    });
  });

  context('with withdrawn LPA', () => {
    it('shows the no options', () => {
      cy.visit('/fixtures?redirect=&progress=withdrawn');

      cy.contains('Property and affairs');
      cy.contains('Sam Smith');
      cy.contains('strong', 'Withdrawn');
      cy.contains('.app-dashboard-card a').should('not.exist');
    });
  });

  context('with registered LPA', () => {
    it('shows the correct options', () => {
      cy.visit('/fixtures?redirect=&progress=registered');

      cy.contains('Property and affairs');
      cy.contains('Sam Smith');
      cy.contains('strong', 'Registered');
      cy.contains('a', 'View signed LPA');
      cy.contains('a', 'Use');
    });
  });

  context('with various roles', () => {
    it('shows all of my LPAs', () => {
      cy.visit('/fixtures/dashboard?asDonor=1&asAttorney=1&asCertificateProvider=1&redirect=/dashboard');

      cy.contains('My LPAs');
      cy.contains('I’m an attorney');
      cy.contains('I’m a certificate provider');
    });
  })
});
