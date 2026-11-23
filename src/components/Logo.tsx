interface LogoProps {
  class?: string;
}

export function Logo(props: LogoProps) {
  return (
    <svg
      viewBox="0 0 74 74"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class={props.class}
      aria-hidden="true"
    >
      <path
        d="M12.5254 24.3464L9.25 18.5L64.75 18.6772L61.479 24.3464H36.9579L24.6531 45.9604L21.3778 40.114L30.3187 24.3464H12.5254Z"
        fill="currentColor"
      />
      <path
        d="M27.5744 50.6553L39.7022 29.0413H58.5727L52.184 40.114H45.9258L48.8201 35.0975H42.8005L33.8242 50.6553L36.9579 56.0867L42.8005 45.9604H48.8201L36.9579 66.6L27.5744 50.6553Z"
        fill="currentColor"
      />
    </svg>
  );
}
