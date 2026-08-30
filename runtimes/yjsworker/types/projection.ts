// InFlightProjection records the current BlockTable projection request for this room.
export type InFlightProjection = {
  connectionId: string;
  connectorChannelId: number;
  projectedSequence: number;
};
